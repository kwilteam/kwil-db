package badger

import (
	"sync"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/kwilteam/kwil-db/pkg/kv"
)

// NewBadgerDB creates a new BadgerDB.
// It takes a path, like path/to/db, where the database will be stored.
func NewBadgerDB(path string) (*BadgerDB, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerDB{db: db}, nil
}

// BadgerDB is a basic threadsafe key-value store.
type BadgerDB struct {
	db *badger.DB

	mu sync.Mutex
}

var _ kv.KVStore = (*BadgerDB)(nil)

// Close closes the underlying database.
func (d *BadgerDB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.db.Close()
}

// Get retrieves a value from the database.
func (d *BadgerDB) Get(key []byte) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var val []byte
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return kv.ErrKeyNotFound
		}
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}

// Set sets a value in the database.
func (d *BadgerDB) Set(key, val []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

// Delete deletes a value from the database.
func (d *BadgerDB) Delete(key []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// BeginTransaction creates a new transaction.
func (d *BadgerDB) BeginTransaction() kv.Transaction {
	d.mu.Lock()
	defer d.mu.Unlock()

	return &Transaction{txn: d.db.NewTransaction(true)}
}

// a Transaction is a basic wrapper around a badger to handle commits and rollbacks
type Transaction struct {
	txn *badger.Txn
}

func (t *Transaction) Commit() error {
	return t.txn.Commit()
}

func (t *Transaction) Discard() {
	t.txn.Discard()
}

func (t *Transaction) Set(key, val []byte) error {
	return t.txn.Set(key, val)
}

func (t *Transaction) Delete(key []byte) error {
	return t.txn.Delete(key)
}

func (t *Transaction) Get(key []byte) ([]byte, error) {
	var val []byte
	item, err := t.txn.Get(key)
	if err != nil {
		return nil, err
	}
	val, err = item.ValueCopy(nil)
	return val, err
}
