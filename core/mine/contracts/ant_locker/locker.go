package ant_locker

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-ipfs/core/mine/sctx"
	"github.com/ipfs/go-ipfs/core/mine/transaction"
	logging "github.com/ipfs/go-log"
	"math/big"
)

var (
	log = logging.Logger("ant_locker")

	lockerABI    = transaction.ParseABIUnchecked(AntLockerABIJson)
	errDecodeABI = errors.New("could not decode abi data")
)

func init() {
	logging.SetLogLevel("ant_locker", "info")
}

type LockInfo struct {
	AntAddress   common.Address
	LockedAmount *big.Int
	LockedAt     *big.Int
}

type Locker struct {
	backend            transaction.Backend
	transactionService transaction.Service
	contractAddress    common.Address
}

func NewLocker(backend transaction.Backend, transactionService transaction.Service, contractAddress common.Address) *Locker {
	return &Locker{
		backend:            backend,
		transactionService: transactionService,
		contractAddress:    contractAddress,
	}
}

func (l *Locker) GetMinLockAmount(ctx context.Context) (*big.Int, error) {
	callData, err := lockerABI.Pack("minLockAmount")
	if err != nil {
		return nil, err
	}

	output, err := l.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &l.contractAddress,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := lockerABI.Unpack("minLockAmount", output)
	if err != nil {
		return nil, err
	}

	if len(results) != 1 {
		return nil, errDecodeABI
	}

	minLock, ok := abi.ConvertType(results[0], new(big.Int)).(*big.Int)
	if !ok || minLock == nil {
		return nil, errDecodeABI
	}
	return minLock, nil
}

func (l *Locker) GetLockInfo(ctx context.Context, antNodeId string) (*LockInfo, error) {
	callData, err := lockerABI.Pack("antLockInfos", antNodeId)
	if err != nil {
		return nil, err
	}

	output, err := l.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &l.contractAddress,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := lockerABI.Unpack("antLockInfos", output)
	if err != nil {
		return nil, err
	}

	//log.Infof("results: %v", results)

	if len(results) != 3 {
		return nil, errDecodeABI
	}

	antAddress, ok := abi.ConvertType(results[0], new(common.Address)).(*common.Address)
	if !ok || antAddress == nil {
		return nil, errDecodeABI
	}
	lockedAmount, ok := abi.ConvertType(results[1], new(big.Int)).(*big.Int)
	if !ok || lockedAmount == nil {
		return nil, errDecodeABI
	}
	lockedAt, ok := abi.ConvertType(results[2], new(big.Int)).(*big.Int)
	if !ok || lockedAt == nil {
		return nil, errDecodeABI
	}
	return &LockInfo{
		AntAddress:   *antAddress,
		LockedAmount: lockedAmount,
		LockedAt:     lockedAt,
	}, nil
}

func (l *Locker) Lock(ctx context.Context, nodeId string, antAddress common.Address) (common.Hash, error) {
	callData, err := lockerABI.Pack("lock", nodeId, antAddress)
	if err != nil {
		log.Errorf("Pack lock: %v", err)
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &l.contractAddress,
		Data:        callData,
		GasPrice:    sctx.GetGasPrice(ctx),
		Value:       big.NewInt(0),
		Description: "lock",
	}

	txHash, err := l.transactionService.Send(ctx, request)
	if err != nil {
		log.Errorf("lock error: %v", err)
		return common.Hash{}, err
	}

	return txHash, nil
}

func (l *Locker) TokenContractAddress(ctx context.Context) (common.Address, error) {
	callData, err := lockerABI.Pack("tokenContract")
	if err != nil {
		return common.Address{}, err
	}

	output, err := l.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &l.contractAddress,
		Data: callData,
	})
	if err != nil {
		return common.Address{}, err
	}

	results, err := lockerABI.Unpack("tokenContract", output)
	if err != nil {
		return common.Address{}, err
	}

	if len(results) != 1 {
		return common.Address{}, errDecodeABI
	}

	erc20Address, ok := abi.ConvertType(results[0], new(common.Address)).(*common.Address)
	if !ok || erc20Address == nil {
		return common.Address{}, errDecodeABI
	}
	return *erc20Address, nil
}
