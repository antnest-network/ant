package chain

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-ipfs/core/mine/transaction"
	"math/big"
)

type Chain interface {
	Backend() transaction.Backend

	TransactionMonitor() transaction.Monitor

	TransactionService() transaction.Service

	LockToken(ctx context.Context, nodeId string, ethAddress common.Address) error

	BNBBalanceOf(ctx context.Context, account common.Address) (*big.Int, error)

	AntzBalanceOf(ctx context.Context, account common.Address) (*big.Int, error)
}
