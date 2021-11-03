package statestore

import "github.com/ipfs/go-datastore"

type StateStore interface {
	Get(key datastore.Key, i interface{}) (err error)
	Put(key datastore.Key, i interface{}) (err error)
	Delete(key datastore.Key) (err error)
	Iterate(prefix string, iterFunc StateIterFunc) (err error)
}

type StateIterFunc func(key string, value []byte) (stop bool, err error)
