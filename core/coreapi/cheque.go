package coreapi

import (
	"context"
	"github.com/ipfs/interface-go-ipfs-core"
)

type ChequeAPI CoreAPI

func (c *ChequeAPI) Get(ctx context.Context, chequebook string) (iface.Cheque, error) {
	return c.chequeManger.GetCheque(ctx, chequebook)
}

func (c *ChequeAPI) List(ctx context.Context) ([]iface.Cheque, error) {
	return c.chequeManger.GetCheques(ctx)
}

func (c *ChequeAPI) CashOut(ctx context.Context, chequebook string) error {
	return c.chequeManger.CashOut(ctx, chequebook)
}

func (c *ChequeAPI) CashOutAll(ctx context.Context) error {
	return c.chequeManger.CashOutAll(ctx)
}
