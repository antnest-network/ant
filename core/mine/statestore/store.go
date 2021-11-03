package statestore

import (
	"encoding"
	"encoding/json"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/ipfs/go-ipfs/repo"
)

type store struct {
	ds repo.Datastore
}

func NewStore(ds repo.Datastore) *store {
	s := store{
		ds: ds,
	}
	return &s
}

func (s *store) Get(key datastore.Key, i interface{}) (err error) {
	data, err := s.ds.Get(key)
	if err != nil {
		return err
	}

	if unmarshaler, ok := i.(encoding.BinaryUnmarshaler); ok {
		return unmarshaler.UnmarshalBinary(data)
	}
	return json.Unmarshal(data, i)
}

func (s *store) Put(key datastore.Key, i interface{}) (err error) {
	var bytes []byte
	if marshaler, ok := i.(encoding.BinaryMarshaler); ok {
		if bytes, err = marshaler.MarshalBinary(); err != nil {
			return err
		}
	} else if bytes, err = json.Marshal(i); err != nil {
		return err
	}

	return s.ds.Put(key, bytes)
}

func (s *store) Delete(key datastore.Key) (err error) {
	return s.ds.Delete(key)
}

func (s *store) Iterate(prefix string, iterFunc StateIterFunc) (err error) {
	q := query.Query{
		Prefix: prefix,
	}
	results, err := s.ds.Query(q)
	if err != nil {
		return err
	}
	es, err := results.Rest()
	if err != nil {
		return err
	}
	for _, r := range es {
		if stop, _ := iterFunc(r.Key, r.Value); stop {
			break
		}
	}
	return nil
}

var _ StateStore = &store{}