package badger

import (
	"context"
	"fmt"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/kwilteam/kwil-db/pkg/kv"
	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
)

// NewBadgerDB creates a new BadgerDB.
// It takes a path, like path/to/db, where the database will be stored.
func NewBadgerDB(ctx context.Context, path string, options *Options) (*BadgerDB, error) {
	b := &BadgerDB{
		logger:     log.NewNoOp(),
		gcInterval: 5 * time.Minute,
	}

	badgerOpts := badger.DefaultOptions(path)

	if options != nil {
		options.applyToBadgerOpts(&badgerOpts)
		options.applyToDB(b)
	}

	badgerOpts.Logger = &badgerLogger{log: b.logger}

	var err error
	b.db, err = badger.Open(badgerOpts)
	if err != nil {
		return nil, err
	}

	go b.runGC(ctx)

	return b, nil
}

// BadgerDB is a basic threadsafe key-value store.
type BadgerDB struct {
	// db is the underlying badger database.
	db *badger.DB

	// mu is a mutex to protect the database.
	mu sync.Mutex

	// gcInterval is the interval at which the database is garbage collected.
	gcInterval time.Duration

	// logger is the logger for the database.
	logger log.Logger
}

var _ kv.KVStore = (*BadgerDB)(nil)

// Close closes the underlying database.
func (d *BadgerDB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Info("closing KV store")

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

func (d *BadgerDB) runGC(ctx context.Context) {
	ticker := time.NewTicker(d.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := d.db.RunValueLogGC(0.5)
			if err == badger.ErrNoRewrite {
				d.logger.Debug("no GC required")
				continue
			}
			if err != nil {
				d.logger.Error("error running GC", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
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

// Options are options for the BadgerDB.
// These get translated into Badger's options.
// We provide this abstraction layer since Badger has a lot of options,
// and I don't want future users of this to worry about all of them.
type Options struct {
	// GuaranteeFSync guarantees that all writes to the wal are fsynced before
	// attempting to be written to the LSM tree.
	GuaranteeFSync bool

	// GarbageCollectionInterval is the interval at which the garbage collector
	// runs.
	GarbageCollectionInterval time.Duration

	// Logger is the logger to use for the database.
	Logger log.Logger
}

// applyToBadgerOpts applies the options to the badger options.
func (o *Options) applyToBadgerOpts(opts *badger.Options) {
	opts.SyncWrites = o.GuaranteeFSync
}

func (o *Options) applyToDB(db *BadgerDB) {
	if o.Logger.L != nil {
		db.logger = o.Logger
	}

	if o.GarbageCollectionInterval != 0 {
		db.gcInterval = o.GarbageCollectionInterval
	}
}

// badgerLogger implements the badger.Logger interface.
type badgerLogger struct {
	log log.Logger
}

func (b *badgerLogger) Debugf(p0 string, p1 ...any) {
	b.log.Debug(fmt.Sprintf(p0, p1...))
}

func (b *badgerLogger) Errorf(p0 string, p1 ...any) {
	b.log.Error(fmt.Sprintf(p0, p1...))
}

func (b *badgerLogger) Infof(p0 string, p1 ...any) {
	b.log.Info(fmt.Sprintf(p0, p1...))
}

func (b *badgerLogger) Warningf(p0 string, p1 ...any) {
	b.log.Warn(fmt.Sprintf(p0, p1...))
}
