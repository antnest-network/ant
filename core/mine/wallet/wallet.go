package wallet

import (
	"crypto/ecdsa"
)

type Wallet interface {
	NewAddress() (*ecdsa.PrivateKey, error)

	Import(privateKeyHex string) (*ecdsa.PrivateKey, error)

	Get(address string) (*ecdsa.PrivateKey, error)

	SetDefaultAddress(address string) error

	GetDefaultAddress() (*ecdsa.PrivateKey, error)

	List() ([]*ecdsa.PrivateKey, error)

	Delete(address string) error
}
