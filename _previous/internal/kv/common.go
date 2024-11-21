package kv

import "errors"

var (
	ErrKeyNotFound = errors.New("key not found")
)

// KVStore is a key-value store that supports transactions.
type KVStore interface {
	KVWriter
	BeginTransaction() Transaction
}

// KVWriter is a subset of the KVStore interface that only allows for
// CRUD operations.
type KVWriter interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Delete(key []byte) error
}

// Transaction is a read-write transaction on a KVStore.
type Transaction interface {
	KVWriter
	Commit() error
	Discard()
}
