package mineservice

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	block2 "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	pin "github.com/ipfs/go-ipfs-pinner"
	"github.com/ipfs/go-ipfs/core/mine/crypto"
	"github.com/ipfs/go-ipfs/core/mine/migration"
	"github.com/ipfs/go-ipfs/core/mine/statestore"
	"github.com/ipfs/go-ipfs/core/mine/transaction"
	"github.com/ipfs/go-ipfs/pkg/xcontext"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	proto "github.com/antnest-network/ant-proto"
	ant_pro "github.com/antnest-network/ant-proto/pb"
	"sync"
	"time"
)

var (
	log = logging.Logger("mineservice")
)

func init() {
	logging.SetLogLevel("mineservice", "info")
}

type MineService struct {
	p2pHost            host.Host
	pinning            pin.Pinner
	messenger          proto.Messenger
	blockService       blockservice.BlockService
	chequeStore        *ChequeStore
	transactionService transaction.Service
	signer             crypto.Signer
	migrator           *migration.Migrator
	queenManager       *QueenManager
	walletAddress      common.Address

	mutex  sync.RWMutex
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func New(h host.Host, messenger proto.Messenger, pinning pin.Pinner, blockService blockservice.BlockService,
	stateStore statestore.StateStore, transactionService transaction.Service, signer crypto.Signer, queens []peer.AddrInfo) *MineService {
	m := &MineService{
		p2pHost:            h,
		messenger:          messenger,
		pinning:            pinning,
		blockService:       blockService,
		signer:             signer,
		chequeStore:        NewChequeStore(stateStore),
		transactionService: transactionService,
		queenManager:       NewQueenManager(queens),
	}

	m.migrator = migration.NewMigrator(m, blockService, pinning)
	m.messenger.SetMessageHandler(proto.ProtocolPingMessage, m.HandlePingMessage)
	m.messenger.SetMessageHandler(proto.ProtocolPushBlockMessage, m.HandlePushBlockMessage)
	m.messenger.SetMessageHandler(proto.ProtocolMigrateBlockMessage, m.HandleMigrateBlockMessage)
	m.messenger.SetMessageHandler(proto.ProtocolCheque, m.HandleChequeMessage)
	m.messenger.SetMessageHandler(proto.ProtocolQueens, m.queenManager.HandleQueenMessage)

	return m
}

func (m *MineService) Start() error {
	m.migrator.Start()

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-ticker.C:
				m.PingQueen(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (m *MineService) Stop() error {
	m.migrator.Stop()

	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
	return nil
}

func (m *MineService) PingQueen(ctx context.Context) {
	p := m.queenManager.GetQueen()
	ping := &ant_pro.Ping{}
	err := xcontext.Do(ctx, func(ctx context.Context) error {
		_, err := m.messenger.Ping(ctx, p.ID, ping)
		return err
	}, xcontext.WithTimeout(time.Second*10))
	if err != nil {
		log.Errorf("failed to ping: %v", err)
		return
	}
	log.Infof("pong from %v", p.ID.String())
}

func (m *MineService) HandlePingMessage(ctx context.Context, from peer.ID, msg interface{}) {
	log.Infof("ping from: %v", from)
	ping, ok := msg.(*ant_pro.Ping)
	if !ok {
		log.Infof("msg type error: %v, %+v", from, msg)
		return
	}
	pong := &ant_pro.Pong{
		Seq: ping.Seq,
	}
	err := xcontext.Do(ctx, func(ctx context.Context) error {
		err := m.messenger.Pong(ctx, from, pong)
		return err
	}, xcontext.WithTimeout(time.Second*5))

	if err != nil {
		log.Warnf("failed to pong: %v", err)
		return
	}
}

func (m *MineService) HandlePushBlockMessage(ctx context.Context, from peer.ID, msg interface{}) {
	req, ok := msg.(*ant_pro.PushBlockReq)
	if !ok {
		log.Infof("msg type error: %v, %+v", from, msg)
		return
	}

	log.Infof("received block from %v, cid: %v", from, req.Cid)
	bcid, err := cid.Decode(req.Cid)
	if err != nil {
		log.Errorf("failed to decode cid: %v", err)
		return
	}
	block, err := block2.NewBlockWithCid(req.Data, bcid)
	if err != nil {
		log.Errorf("failed to NewBlockWithCid: %v", err)
		return
	}

	resp := ant_pro.PushBlockResp{
		Seq:  req.Seq,
		Code: proto.Success,
	}

	err = m.blockService.AddBlock(block)
	if err != nil {
		log.Errorf("failed to AddBlock: %v", err)
		resp.ErrString = err.Error()
		resp.Code = proto.Failure
	}
	m.pinning.PinWithMode(bcid, pin.Direct)

	err = xcontext.Do(ctx, func(ctx context.Context) error {
		return m.messenger.RespondPushBlock(ctx, from, &resp)
	}, xcontext.WithTimeout(time.Second*10), xcontext.WithTryCount(3))
	if err != nil {
		log.Errorf("failed to RespondPushBlock: %v", err)
		return
	}
}

func (m *MineService) HandleMigrateBlockMessage(ctx context.Context, from peer.ID, msg interface{}) {
	req, ok := msg.(*ant_pro.MigrateBlockReq)
	if !ok {
		log.Infof("msg type error: %v, %+v", from, msg)
		return
	}
	log.Infof("received migrate message from %v", from)
	fromAnt := peer.ID(req.FromAnt)
	for _, block := range req.Cids {
		bcid, err := cid.Decode(block)
		if err != nil {
			log.Errorf("failed to decode cid: %v", err)
			continue
		}
		m.migrator.AsyncMigrate(fromAnt, bcid)
	}
	resp := ant_pro.MigrateBlockResp{
		Seq:  req.Seq,
		Code: proto.Success,
	}
	err := xcontext.Do(ctx, func(ctx context.Context) error {
		return m.messenger.RespondMigrateBlock(ctx, from, &resp)
	}, xcontext.WithTimeout(time.Second*10), xcontext.WithTryCount(3))

	if err != nil {
		log.Errorf("failed to RespondMigrateBlock: %v", err)
		return
	}
}

func (m *MineService) OnMigrateBlockDone(ant peer.ID, bcid cid.Cid, code int) error {
	msg := &ant_pro.MigrateBlockResult{
		FromAnt: string(ant),
		Blocks:  make([]*ant_pro.MigrateBlockResult_Block, 0),
	}
	msg.Blocks = append(msg.Blocks, &ant_pro.MigrateBlockResult_Block{
		Cid:  bcid.String(),
		Code: int32(code),
	})

	to := m.queenManager.GetQueen()
	err := xcontext.Do(context.Background(), func(ctx context.Context) error {
		return m.messenger.SendMigrateBlockResult(ctx, to.ID, msg)
	}, xcontext.WithTimeout(time.Second*10), xcontext.WithTryCount(3))

	if err != nil {
		log.Errorf("failed to send migrate block result: %v", err)
		return err
	}
	return nil
}

func (m *MineService) HandleChequeMessage(ctx context.Context, from peer.ID, msg interface{}) {
	cheque, ok := msg.(*ant_pro.Cheque)
	if !ok {
		log.Infof("msg type error: %v, %+v", from, msg)
		return
	}
	log.Infof("received cheque from %v, %v", from, cheque.String())
	err := m.chequeStore.SaveCheque(cheque)
	if err != nil {
		log.Errorf("failed to save cheque: %v", err)
		return
	}
}
