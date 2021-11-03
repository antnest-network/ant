package migration

import (
	"context"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	pin "github.com/ipfs/go-ipfs-pinner"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	proto "github.com/antnest-network/ant-proto"
	"sync"
	"time"
)

var log = logging.Logger("migration")

func init() {
	logging.SetLogLevel("migration", "info")
}

type Notifier interface {
	OnMigrateBlockDone(ant peer.ID, block cid.Cid, code int) error
}

type Block struct {
	fromAnt peer.ID
	cid     cid.Cid
	Code    int
}

type Migrator struct {
	notifier      Notifier
	blockService  blockservice.BlockService
	pinning       pin.Pinner
	blocksRequest chan *Block
	blocksDone    chan *Block
	wg            sync.WaitGroup
	cancel        context.CancelFunc
}

func NewMigrator(n Notifier, blockService blockservice.BlockService, pinning pin.Pinner) *Migrator {
	return &Migrator{
		notifier:      n,
		blockService:  blockService,
		pinning:       pinning,
		blocksRequest: make(chan *Block, 1000),
		blocksDone:    make(chan *Block, 1000),
	}
}

func (m *Migrator) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case block := <-m.blocksRequest:
				m.Migrate(ctx, block)
			case block := <-m.blocksDone:
				m.notifier.OnMigrateBlockDone(block.fromAnt, block.cid, block.Code)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (m *Migrator) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
	return nil
}

func (m *Migrator) AsyncMigrate(fromAnt peer.ID, bcid cid.Cid) {
	m.blocksRequest <- &Block{
		fromAnt: fromAnt,
		cid:     bcid,
	}
}

func (m *Migrator) Migrate(ctx context.Context, block *Block) error {
	if has, _ := m.blockService.Blockstore().Has(block.cid); has {
		m.pinning.PinWithMode(block.cid, pin.Direct)
		m.blocksDone <- block
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()
	_, err := m.blockService.GetBlock(ctx, block.cid)
	if err != nil {
		log.Errorf("failed to get block %v: %v", block.cid.String(), err)
		block.Code = proto.Failure
		m.blocksDone <- block
		return err
	}
	m.pinning.PinWithMode(block.cid, pin.Direct)
	m.blocksDone <- block
	return nil
}
