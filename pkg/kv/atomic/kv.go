package atomic

import (
	"sync"

	"github.com/kwilteam/kwil-db/pkg/kv"
	"github.com/kwilteam/kwil-db/pkg/sessions"
)

// NewAtomicKV creates a new atomic key value store.
// It is compatible with Kwil's 2pc protocol.
func NewAtomicKV(store kv.KVStore) (*AtomicKV, error) {
	// return an interface for newPrefix
	return &AtomicKV{
		baseKV: &baseKV{
			db:              store,
			inSession:       false,
			uncommittedData: []*keyValue{},
		},
		prefix: []byte{},
	}, nil
}

// AtomicKV is a threadsafe, linearizable key/value store.
type AtomicKV struct {
	*baseKV
	// prefix is the prefix for this store.
	prefix []byte
}

type baseKV struct {
	mu sync.RWMutex

	// db is the underlying badger database.
	// you should not write to this directly.
	db kv.KVStore

	// inSession is true if we are currently in a session.
	// if in a session, data is not written to the database,
	// but instead stored to be written when the session is committed.
	inSession bool

	//uncommittedData is a map of keys to values that have not yet been committed.
	uncommittedData []*keyValue

	// currentTxn is the current transaction.
	currentTx kv.Transaction
}

var _ sessions.Committable = &AtomicKV{}
var _ kv.KVWriter = &AtomicKV{}

// Write writes a key/value pair to the database.
func (k *AtomicKV) Set(key []byte, value []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.inSession {
		k.uncommittedData = append(k.uncommittedData, &keyValue{
			Operation: kvOperationSet,
			Key:       append(k.prefix, key...),
			Value:     value,
		})
		return nil
	}

	return k.db.Set(append(k.prefix, key...), value)
}

// Read reads a key from the database.
func (k *AtomicKV) Get(key []byte) ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.db.Get(append(k.prefix, key...))
}

// Delete deletes a key from the database.
func (k *AtomicKV) Delete(key []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.inSession {
		k.uncommittedData = append(k.uncommittedData, &keyValue{
			Operation: kvOperationDelete,
			Key:       append(k.prefix, key...),
		})
		return nil
	}

	return k.db.Delete(append(k.prefix, key...))
}

// NewPrefix creates a new writer with a prefix for the current store.
func (k *AtomicKV) NewPrefix(prefix []byte) *AtomicKV {
	return &AtomicKV{
		baseKV: &baseKV{
			db:              k.db,
			inSession:       k.inSession,
			uncommittedData: k.uncommittedData,
		},
		prefix: append(k.prefix, prefix...),
	}
}
