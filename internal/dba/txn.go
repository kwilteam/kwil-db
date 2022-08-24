package dba

import (
	"github.com/dgraph-io/badger/v3"
)

type txn struct {
	btx *badger.Txn
}

func (t txn) Get(key []byte) ([]byte, error) {
	bItem, err := t.btx.Get(key)
	if err != nil {
		return nil, err
	}
	val, err := bItem.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (t txn) Set(key, val []byte) error {
	return t.btx.Set(key, val)
}

func (t txn) Discard() {
	t.btx.Discard()
}

func (t txn) Commit() error {
	return t.btx.Commit()
}
