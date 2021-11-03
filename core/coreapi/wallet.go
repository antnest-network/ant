package coreapi

import (
	"context"
	"crypto/ecdsa"
)

type WalletAPI CoreAPI

func (w *WalletAPI) NewAddress(ctx context.Context) (*ecdsa.PrivateKey, error) {
	return w.wallet.NewAddress()
}

func (w *WalletAPI) Import(ctx context.Context, privateKeyHex string) (*ecdsa.PrivateKey, error) {
	return w.wallet.Import(privateKeyHex)
}

func (w *WalletAPI) Get(ctx context.Context, address string) (*ecdsa.PrivateKey, error) {
	return w.wallet.Get(address)
}

func (w *WalletAPI) SetDefaultAddress(ctx context.Context, address string) error {
	return w.wallet.SetDefaultAddress(address)
}

func (w *WalletAPI) GetDefaultAddress(ctx context.Context) (*ecdsa.PrivateKey, error) {
	return w.wallet.GetDefaultAddress()
}

func (w *WalletAPI) List(ctx context.Context) ([]*ecdsa.PrivateKey, error) {
	return w.wallet.List()
}

func (w *WalletAPI) Delete(ctx context.Context, address string) error {
	return w.wallet.Delete(address)
}
