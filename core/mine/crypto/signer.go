// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crypto

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ipfs/go-ipfs/core/mine/crypto/eip712"
)

var (
	ErrInvalidLength = errors.New("invalid signature length")
)

type Signer interface {
	// Sign signs data with ethereum prefix (eip191 type 0x45).
	Sign(data []byte) ([]byte, error)
	// SignTx signs an ethereum transaction.
	SignTx(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error)
	// SignTypedData signs data according to eip712.
	SignTypedData(typedData *eip712.TypedData) ([]byte, error)
	// PublicKey returns the public key this signer uses.
	PublicKey() (*ecdsa.PublicKey, error)
	// EthereumAddress returns the ethereum address this signer uses.
	EthereumAddress() (common.Address, error)
}

// addEthereumPrefix adds the ethereum prefix to the data.
func addEthereumPrefix(data []byte) []byte {
	return []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))
}

// hashWithEthereumPrefix returns the hash that should be signed for the given data.
func hashWithEthereumPrefix(data []byte) ([]byte, error) {
	return LegacyKeccak256(addEthereumPrefix(data))
}

// Recover verifies signature with the data base provided.
// It is using `btcec.RecoverCompact` function.
func Recover(signature, data []byte) (*ecdsa.PublicKey, error) {
	if len(signature) != 65 {
		return nil, ErrInvalidLength
	}
	// Convert to btcec input format with 'recovery id' v at the beginning.
	btcsig := make([]byte, 65)
	btcsig[0] = signature[64]
	copy(btcsig[1:], signature)

	hash, err := hashWithEthereumPrefix(data)
	if err != nil {
		return nil, err
	}

	p, _, err := btcec.RecoverCompact(btcec.S256(), btcsig, hash)
	return (*ecdsa.PublicKey)(p), err
}

// RecoverEIP712 recovers the public key for eip712 signed data.
func RecoverEIP712(signature []byte, data *eip712.TypedData) (*ecdsa.PublicKey, error) {
	if len(signature) != 65 {
		return nil, errors.New("invalid length")
	}
	// Convert to btcec input format with 'recovery id' v at the beginning.
	btcsig := make([]byte, 65)
	btcsig[0] = signature[64]
	copy(btcsig[1:], signature)

	rawData, err := eip712.EncodeForSigning(data)
	if err != nil {
		return nil, err
	}

	sighash, err := LegacyKeccak256(rawData)
	if err != nil {
		return nil, err
	}

	p, _, err := btcec.RecoverCompact(btcec.S256(), btcsig, sighash)
	return (*ecdsa.PublicKey)(p), err
}
