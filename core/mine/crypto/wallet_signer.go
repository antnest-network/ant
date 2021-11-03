package crypto

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ipfs/go-ipfs/core/mine/crypto/eip712"
	"github.com/ipfs/go-ipfs/core/mine/wallet"
	"math/big"
)

type walletSigner struct {
	w wallet.Wallet
}

func NewWalletSigner(w wallet.Wallet) Signer {
	return &walletSigner{
		w: w,
	}
}

// PublicKey returns the public key this signer uses.
func (s *walletSigner) PublicKey() (*ecdsa.PublicKey, error) {
	signer, err := s.signer()
	if err != nil {
		return nil, err
	}
	return signer.PublicKey()
}

// Sign signs data with ethereum prefix (eip191 type 0x45).
func (s *walletSigner) Sign(data []byte) (signature []byte, err error) {
	signer, err := s.signer()
	if err != nil {
		return nil, err
	}
	return signer.Sign(data)
}

// SignTx signs an ethereum transaction.
func (s *walletSigner) SignTx(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	signer, err := s.signer()
	if err != nil {
		return nil, err
	}
	return signer.SignTx(transaction, chainID)
}

// EthereumAddress returns the ethereum address this signer uses.
func (s *walletSigner) EthereumAddress() (common.Address, error) {
	signer, err := s.signer()
	if err != nil {
		return common.Address{}, err
	}
	return signer.EthereumAddress()
}

// SignTypedData signs data according to eip712.
func (s *walletSigner) SignTypedData(typedData *eip712.TypedData) ([]byte, error) {
	signer, err := s.signer()
	if err != nil {
		return nil, err
	}
	return signer.SignTypedData(typedData)
}

func (s *walletSigner) signer() (Signer, error) {
	privateKey, err := s.w.GetDefaultAddress()
	if err != nil {
		return nil, err
	}
	return NewDefaultSigner(privateKey), nil
}

var _ Signer = &walletSigner{}