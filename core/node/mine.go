package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-config"
	pin "github.com/ipfs/go-ipfs-pinner"
	"github.com/ipfs/go-ipfs/core/mine/chain"
	"github.com/ipfs/go-ipfs/core/mine/crypto"
	"github.com/ipfs/go-ipfs/core/mine/mineservice"
	"github.com/ipfs/go-ipfs/core/mine/statestore"
	"github.com/ipfs/go-ipfs/core/mine/wallet"
	"github.com/ipfs/go-ipfs/core/mine/wallet/localwallet"
	"github.com/ipfs/go-ipfs/repo"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/antnest-network/ant-proto"
	"go.uber.org/fx"
	"time"
)

const blocktime = time.Second * 3

func NewStateStore(repo repo.Repo) statestore.StateStore {
	return statestore.NewStore(repo.Datastore())
}

func NewLocalWallet(store statestore.StateStore) (wallet.Wallet, error) {
	w := localwallet.NewLocalWallet(store)
	_, err := w.GetDefaultAddress()
	if err == datastore.ErrNotFound {
		_, err := w.NewAddress()
		return w, err
	}
	return w, nil
}

func NewSigner(w wallet.Wallet) crypto.Signer {
	return crypto.NewWalletSigner(w)
}

func NewChain(signer crypto.Signer, stateStore statestore.StateStore, cfg *config.Config) (chain.Chain, error) {
	ethAddress, err := signer.EthereumAddress()
	if err != nil {
		return nil, err
	}
	if !common.IsHexAddress(cfg.Ant.Chain.LockerContract) {
		return nil, errors.New(fmt.Sprintf("LockerContract is error: %v", cfg.Ant.Chain.LockerContract))
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	ch, err := chain.NewChain(ctx, ethAddress, stateStore, signer, cfg.Ant.Chain.Endpoint, common.HexToAddress(cfg.Ant.Chain.LockerContract))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to InitChain: %v", err))
	}
	return ch, nil
}

func NewChequeManager(stateStore statestore.StateStore, chx chain.Chain) *mineservice.ChequeManager {
	return mineservice.NewChequeManager(mineservice.NewChequeStore(stateStore), chx.TransactionService())
}

func NewMineService(lc fx.Lifecycle, h host.Host, messenger proto.Messenger, pinning pin.Pinner,
	blockService blockservice.BlockService, signer crypto.Signer, chx chain.Chain,
	stateStore statestore.StateStore, cfg *config.Config) (*mineservice.MineService, error) {
	ethAddress, _ := signer.EthereumAddress()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	if !common.IsHexAddress(cfg.Ant.Chain.LockerContract) {
		return nil, errors.New(fmt.Sprintf("LockerContract is error: %v", cfg.Ant.Chain.LockerContract))
	}
	err := chx.LockToken(ctx, cfg.Identity.PeerID, ethAddress)
	if err != nil {
		return nil, err
	}

	queens, err := config.ParseBootstrapPeers(cfg.Ant.QueenAddresses)
	if err != nil {
		return nil, errors.New("failed to parse queen address")
	}
	ms := mineservice.New(h, messenger, pinning, blockService, stateStore, chx.TransactionService(), signer, queens)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return ms.Start()
		},
		OnStop: func(ctx context.Context) error {
			return ms.Stop()
		},
	})
	return ms, nil
}
