package chequebook

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
	log = logging.Logger("chequebookcontract")

	chequebookABI = transaction.ParseABIUnchecked(ERC20SimpleSwapJson)
	errDecodeABI  = errors.New("could not decode abi data")
)

func init() {
	logging.SetLogLevel("chequebookcontract", "info")
}

type ChequebookContract struct {
	transactionService transaction.Service
}

func NewChequebookContract(transactionService transaction.Service) *ChequebookContract {
	return &ChequebookContract{
		transactionService: transactionService,
	}
}

func (c *ChequebookContract) PaidOut(ctx context.Context, chequebook, address common.Address) (*big.Int, error) {
	callData, err := chequebookABI.Pack("paidOut", address)
	if err != nil {
		return nil, err
	}

	output, err := c.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &chequebook,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := chequebookABI.Unpack("paidOut", output)
	if err != nil {
		return nil, err
	}

	return abi.ConvertType(results[0], new(big.Int)).(*big.Int), nil
}

func (c *ChequebookContract) CashCheque(ctx context.Context, chequebook, recipient common.Address,
	cumulativePayout *big.Int, signature []byte) (common.Hash, error) {
	callData, err := chequebookABI.Pack("cashCheque", recipient, cumulativePayout, signature)
	if err != nil {
		return common.Hash{}, err
	}
	lim := sctx.GetGasLimit(ctx)
	if lim == 0 {
		// fix for out of gas errors
		lim = 300000
	}
	request := &transaction.TxRequest{
		To:          &chequebook,
		Data:        callData,
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    lim,
		Value:       big.NewInt(0),
		Description: "cheque cashout",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}
	return txHash, err
}
