package chain

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-ipfs/core/mine/contracts/ant_locker"
	"github.com/ipfs/go-ipfs/core/mine/contracts/erc20"
	"github.com/ipfs/go-ipfs/core/mine/crypto"
	"github.com/ipfs/go-ipfs/core/mine/statestore"
	"github.com/ipfs/go-ipfs/core/mine/transaction"
	logging "github.com/ipfs/go-log"
	"math/big"
	"time"
)

var log = logging.Logger("chain")

func init() {
	logging.SetLogLevel("chain", "info")
}

const (
	maxDelay          = 1 * time.Minute
	cancellationDepth = 6
	blocktime         = time.Second * 3
)

var (
	balanceCheckBackoffDuration = 3 * time.Second
	balanceCheckMaxRetries      = 10
)

type BlockChain struct {
	lockerContract     common.Address
	tokenContract      common.Address
	ethClient          *ethclient.Client
	transactionMonitor transaction.Monitor
	transactionService transaction.Service
}

func NewChain(ctx context.Context,
	ethAddress common.Address,
	stateStore statestore.StateStore,
	signer crypto.Signer,
	endpoint string,
	lockerContract common.Address) (*BlockChain, error) {
	backend, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("dial eth client: %w", err)
	}
	chainID, err := backend.ChainID(ctx)
	if err != nil {
		log.Infof("could not connect to backend at %v", endpoint)
		return nil, fmt.Errorf("get chain id: %w", err)
	}
	transactionMonitor := transaction.NewMonitor(backend, ethAddress, blocktime, cancellationDepth)
	transactionService, err := transaction.NewService(backend, signer, stateStore, chainID, transactionMonitor)
	if err != nil {
		return nil, fmt.Errorf("new transaction service: %w", err)
	}
	locker := ant_locker.NewLocker(backend, transactionService, lockerContract)
	tokenContract, err := locker.TokenContractAddress(ctx)
	if err != nil {
		log.Errorf("failed to get token contract address: %v", err)
		return nil, err
	}

	return &BlockChain{
		ethClient:          backend,
		lockerContract:     lockerContract,
		tokenContract:      tokenContract,
		transactionMonitor: transactionMonitor,
		transactionService: transactionService,
	}, nil
}

func (c *BlockChain) Backend() transaction.Backend {
	return c.ethClient
}

func (c *BlockChain) TransactionMonitor() transaction.Monitor {
	return c.transactionMonitor
}

func (c *BlockChain) TransactionService() transaction.Service {
	return c.transactionService
}

func (c *BlockChain) LockToken(
	ctx context.Context,
	nodeId string,
	ethAddress common.Address,
) error {
	log.Infof("eth address: %v", ethAddress)

	locker := ant_locker.NewLocker(c.ethClient, c.transactionService, c.lockerContract)

	lockerInfo, err := locker.GetLockInfo(ctx, nodeId)
	if err == nil && lockerInfo.LockedAmount.Cmp(big.NewInt(0)) > 0 {
		log.Infof("lock token: %v", lockerInfo.LockedAmount.String())
		return nil
	}
	if err != nil {
		log.Errorf("tail to get GetLockInfo: %v", err)
	}

	if err := checkEthBalance(ctx, c.ethClient, ethAddress); err != nil {
		return err
	}

	tokenContract, err := locker.TokenContractAddress(ctx)
	if err != nil {
		log.Errorf("failed to get token contract address: %v", err)
		return err
	}
	lockAmount, err := locker.GetMinLockAmount(ctx)
	if err != nil {
		log.Errorf("failed to get token contract address: %v", err)
		return err
	}
	erc20Token := erc20.New(c.ethClient, c.transactionService, tokenContract)
	if err := checkTokenBalance(ctx, erc20Token, ethAddress, lockAmount); err != nil {
		return err
	}
	tx, err := erc20Token.Approve(ctx, c.lockerContract, lockAmount)
	if err != nil {
		log.Errorf("failed to approve: %v", err)
		return err
	}
	receipt, err := c.transactionService.WaitForReceipt(ctx, tx)
	if err != nil {
		log.Errorf("failed to approve: %v", err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Errorf("failed to approve: %v", receipt.Status)
		return errors.New(fmt.Sprintf("receipt=%v", receipt))
	}

	tx, err = locker.Lock(ctx, nodeId, ethAddress)
	if err != nil {
		log.Errorf("failed to log: %v", err)
		return err
	}

	receipt, err = c.transactionService.WaitForReceipt(ctx, tx)
	if err != nil {
		log.Errorf("failed to token token: %v", err)
		return err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Errorf("failed to lock token: %v", receipt.Status)
		return errors.New(fmt.Sprintf("receipt=%v", receipt))
	}
	return nil
}

func (c *BlockChain) BNBBalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	return c.ethClient.BalanceAt(ctx, account, nil)
}

func (c *BlockChain) AntzBalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	erc20Token := erc20.New(c.ethClient, c.transactionService, c.tokenContract)
	return erc20Token.BalanceOf(ctx, account)
}
