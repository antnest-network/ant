package localwallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core/mine/crypto"
	"github.com/ipfs/go-ipfs/core/mine/statestore"
	"github.com/ipfs/go-ipfs/core/mine/wallet"
	logging "github.com/ipfs/go-log"
)

var (
	walletKeyPrefix   = "/wallet/list/"
	defaultAddressKey = datastore.NewKey("/wallet/default")
	log               = logging.Logger("localwallet")
)

func init() {
	logging.SetLogLevel("localwallet", "info")
}

func walletKey(address string) datastore.Key {
	return datastore.NewKey(walletKeyPrefix + address)
}

type LocalWallet struct {
	store statestore.StateStore
}

func NewLocalWallet(store statestore.StateStore) *LocalWallet {
	return &LocalWallet{
		store: store,
	}
}

func (w *LocalWallet) NewAddress() (*ecdsa.PrivateKey, error) {
	privateKey, err := crypto.GenerateSecp256k1Key()
	if err != nil {
		return privateKey, err
	}
	err = w.savePrivateKey(privateKey)
	return privateKey, err
}

func (w *LocalWallet) Import(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	privKey, err := DecodePrivateKey(privateKeyHex)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to DecodeSecp256k1PrivateKey: %v", err))
	}
	err = w.savePrivateKey(privKey)
	return privKey, err
}

func (w *LocalWallet) Get(address string) (*ecdsa.PrivateKey, error) {
	addr := common.HexToAddress(address)
	var privateKey string
	err := w.store.Get(walletKey(addr.String()), &privateKey)
	if err != nil {
		return nil, err
	}
	return DecodePrivateKey(privateKey)
}

func (w *LocalWallet) SetDefaultAddress(address string) error {
	privateKey, err := w.Get(address)
	if err != nil {
		return err
	}
	err = w.store.Put(defaultAddressKey, EncodePrivateKey(privateKey))
	return err
}

func (w *LocalWallet) GetDefaultAddress() (*ecdsa.PrivateKey, error) {
	var privateKey string
	err := w.store.Get(defaultAddressKey, &privateKey)
	if err != nil {
		return nil, err
	}
	return DecodePrivateKey(privateKey)
}

func (w *LocalWallet) List() ([]*ecdsa.PrivateKey, error) {
	var list []*ecdsa.PrivateKey
	w.store.Iterate(walletKeyPrefix, func(key string, value []byte) (stop bool, err error) {
		var val string
		err = json.Unmarshal(value, &val)
		if err != nil {
			log.Errorf("iterate private key: %v", err)
			return false, err
		}
		pk, err := DecodePrivateKey(val)
		if err != nil {
			log.Errorf("iterate private key: %v", err)
			return false, err
		}
		list = append(list, pk)
		return false, nil
	})
	return list, nil
}

func (w *LocalWallet) Delete(address string) error {
	addr := common.HexToAddress(address)
	if defaultAddress, err := w.GetDefaultAddress(); err != nil {
		return err
	} else if crypto.EthereumAddress(defaultAddress.PublicKey) == addr.String() {
		return errors.New("cannot ont delete default address")
	}
	err := w.store.Delete(walletKey(addr.String()))
	return err
}

func (w *LocalWallet) savePrivateKey(privateKey *ecdsa.PrivateKey) error {
	err := w.store.Put(walletKey(crypto.EthereumAddress(privateKey.PublicKey)), EncodePrivateKey(privateKey))
	if err != nil {
		return err
	}
	if _, err := w.GetDefaultAddress(); err != nil {
		err = w.store.Put(defaultAddressKey, EncodePrivateKey(privateKey))
		if err != nil {
			return err
		}
	}
	return nil
}

func EncodePrivateKey(key *ecdsa.PrivateKey) string {
	return hex.EncodeToString(crypto.EncodeSecp256k1PrivateKey(key))
}

func DecodePrivateKey(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	data, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to decode privkey, %v", err))
	}
	return crypto.DecodeSecp256k1PrivateKey(data)
}

var _ wallet.Wallet = &LocalWallet{}
