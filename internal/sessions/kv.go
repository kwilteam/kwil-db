package sessions

import (
	kvTypes "github.com/kwilteam/kwil-db/internal/kv"
)

type KV interface {
	Get(key []byte) ([]byte, error)
	Set(key []byte, value []byte) error
	Delete(key []byte) error
}

// setCurrentKey sets the current key in the given connection.
func setCurrentKey(kv KV, idempotencyKey []byte) error {
	return kv.Set(currentKeyKey, idempotencyKey)
}

// deleteCurrentKey deletes the current key in the given connection.
func deleteCurrentKey(kv KV) error {
	return kv.Delete(currentKeyKey)
}

// getCurrentKey gets the current key in the given connection.
// it can return nil if there is no current key.
func getCurrentKey(kv KV) ([]byte, error) {
	val, err := kv.Get(currentKeyKey)
	if err == kvTypes.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return val, nil
}

var (
	currentKeyKey = []byte("currentKey")
)
