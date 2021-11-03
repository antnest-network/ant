package coreapi

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ChainAPI CoreAPI

func (c *ChainAPI) GetBnbBalanceOf(ctx context.Context, account string) (*big.Int, error) {
	return c.chain.BNBBalanceOf(ctx, common.HexToAddress(account))
}

func (c *ChainAPI) GetAntzBalanceOf(ctx context.Context, account string) (*big.Int, error) {
	return c.chain.AntzBalanceOf(ctx, common.HexToAddress(account))
}
