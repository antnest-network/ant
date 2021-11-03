package mineservice

import (
	"encoding/json"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs/core/mine/statestore"
	ant_pro "github.com/antnest-network/ant-proto/pb"
)

const (
	chequePrefix = "/cheque/"
)

func ChequeKey(chequebook string) datastore.Key {
	return datastore.NewKey(chequePrefix + chequebook)
}

type ChequeStore struct {
	stateStore statestore.StateStore
}

func NewChequeStore(stateStore statestore.StateStore) *ChequeStore {
	return &ChequeStore{
		stateStore: stateStore,
	}
}

func (c *ChequeStore) SaveCheque(cheque *ant_pro.Cheque) error {
	key := ChequeKey(cheque.Chequebook)
	return c.stateStore.Put(key, cheque)
}

func (c *ChequeStore) GetCheque(chequebook string) (*ant_pro.Cheque, error) {
	key := ChequeKey(chequebook)
	val := &ant_pro.Cheque{}
	err := c.stateStore.Get(key, val)
	return val, err
}

func (c *ChequeStore) GetCheques() ([]*ant_pro.Cheque, error) {
	var list []*ant_pro.Cheque
	c.stateStore.Iterate(chequePrefix, func(key string, value []byte) (stop bool, err error) {
		cheque := ant_pro.Cheque{}
		err = json.Unmarshal(value, &cheque)
		if err != nil {
			log.Errorf("failed to Unmarshal: %v", err)
			return false, err
		}
		list = append(list, &cheque)
		return false, nil
	})
	//log.Infof("Cheques: %v", list)
	return list, nil
}
