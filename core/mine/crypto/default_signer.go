package crypto

import (
	"crypto/ecdsa"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ipfs/go-ipfs/core/mine/crypto/eip712"
	"math/big"
)

type defaultSigner struct {
	key *ecdsa.PrivateKey
}

func NewDefaultSigner(key *ecdsa.PrivateKey) Signer {
	return &defaultSigner{
		key: key,
	}
}

// PublicKey returns the public key this signer uses.
func (d *defaultSigner) PublicKey() (*ecdsa.PublicKey, error) {
	return &d.key.PublicKey, nil
}

// Sign signs data with ethereum prefix (eip191 type 0x45).
func (d *defaultSigner) Sign(data []byte) (signature []byte, err error) {
	hash, err := hashWithEthereumPrefix(data)
	if err != nil {
		return nil, err
	}

	return d.sign(hash, true)
}

// SignTx signs an ethereum transaction.
func (d *defaultSigner) SignTx(transaction *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	txSigner := types.NewEIP155Signer(chainID)
	hash := txSigner.Hash(transaction).Bytes()
	// isCompressedKey is false here so we get the expected v value (27 or 28)
	signature, err := d.sign(hash, false)
	if err != nil {
		return nil, err
	}

	// v value needs to be adjusted by 27 as transaction.WithSignature expects it to be 0 or 1
	signature[64] -= 27
	return transaction.WithSignature(txSigner, signature)
}

// EthereumAddress returns the ethereum address this signer uses.
func (d *defaultSigner) EthereumAddress() (common.Address, error) {
	publicKey, err := d.PublicKey()
	if err != nil {
		return common.Address{}, err
	}
	eth, err := NewEthereumAddress(*publicKey)
	if err != nil {
		return common.Address{}, err
	}
	var ethAddress common.Address
	copy(ethAddress[:], eth)
	return ethAddress, nil
}

// SignTypedData signs data according to eip712.
func (d *defaultSigner) SignTypedData(typedData *eip712.TypedData) ([]byte, error) {
	rawData, err := eip712.EncodeForSigning(typedData)
	if err != nil {
		return nil, err
	}

	sighash, err := LegacyKeccak256(rawData)
	if err != nil {
		return nil, err
	}

	return d.sign(sighash, false)
}

// sign the provided hash and convert it to the ethereum (r,s,v) format.
func (d *defaultSigner) sign(sighash []byte, isCompressedKey bool) ([]byte, error) {
	signature, err := btcec.SignCompact(btcec.S256(), (*btcec.PrivateKey)(d.key), sighash, false)
	if err != nil {
		return nil, err
	}

	// Convert to Ethereum signature format with 'recovery id' v at the end.
	v := signature[0]
	copy(signature, signature[1:])
	signature[64] = v
	return signature, nil
}

var _ Signer = &defaultSigner{}