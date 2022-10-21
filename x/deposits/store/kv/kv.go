package kv

import (
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*
	I separated this out since this will likely have to be switched with a different kv store
*/

const (
	// Default BadgerDB discardRatio. It represents the discard ratio for the
	// BadgerDB GC.
	//
	// Ref: https://godoc.org/github.com/dgraph-io/badger#DB.RunValueLogGC
	badgerDiscardRatio = 0.5

	// Default BadgerDB GC interval
	badgerGCInterval = 10 * time.Minute
)

type badgerDB struct {
	db  *badger.DB
	log zerolog.Logger
	mu  sync.Mutex
}

type KVStore interface {
	Close() error
	Get(key []byte) ([]byte, error)
	Set(key, val []byte) error
	RunGC()
	Keys(prefix []byte) ([][]byte, error)
	Delete(key []byte) error
	NewTransaction(writeable bool) *badger.Txn
	DeleteByPrefix(prefix []byte) error
	Exists(key []byte) (bool, error)
	PrintAll()
}

func New(path string) (KVStore, error) {
	// create logger
	logger := log.With().Str("module", "store").Logger()

	// create badger db
	opts := badger.DefaultOptions(path)
	opts.SyncWrites = true
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &badgerDB{
		db:  db,
		log: logger,
	}, nil
}

func (db *badgerDB) Close() error {
	return db.db.Close()
}

func (db *badgerDB) Get(key []byte) ([]byte, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var val []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotFound
			} else {
				return err
			}
		}
		val, err = item.ValueCopy(nil)
		return err
	})

	return val, err
}

func (db *badgerDB) Set(key, val []byte) error {

	return db.db.Update(func(txn *badger.Txn) error {
		db.mu.Lock()
		defer db.mu.Unlock()
		return txn.Set(key, val)
	})
}

func (db *badgerDB) RunGC() {
	ticker := time.NewTicker(badgerGCInterval)
	for range ticker.C {
		<-ticker.C
		err := db.db.RunValueLogGC(badgerDiscardRatio)
		if err != nil {
			// don't report error when GC didn't result in any cleanup
			if err == badger.ErrNoRewrite {
				db.log.Debug().Err(err).Msg("no cleanup done")
			} else {
				db.log.Error().Err(err).Msg("failed to run GC")
			}
		}
	}
}

func (db *badgerDB) Delete(key []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (db *badgerDB) NewTransaction(writeable bool) *badger.Txn {
	return db.db.NewTransaction(writeable)
}

func (db *badgerDB) DeleteByPrefix(prefix []byte) error {

	db.mu.Lock()
	defer db.mu.Unlock()

	deleteKeys := func(keysForDelete [][]byte) error {
		if err := db.db.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000
	_ = db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}

		return nil
	})
	return nil
}

func (db *badgerDB) GetAllByPrefix(pref []byte) ([][]byte, [][]byte, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	var keys [][]byte
	var vals [][]byte
	err := db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(pref); it.ValidForPrefix(pref); it.Next() {
			key := it.Item().KeyCopy(nil)
			keys = append(keys, key)
			val, err := it.Item().ValueCopy(nil)
			if err != nil {
				fmt.Println("error finding value for key:", string(key))
				fmt.Println("will continue printing other keys...")
			}
			vals = append(vals, val)
		}
		return nil
	})
	return keys, vals, err
}

func (db *badgerDB) PrintAll() {

	db.mu.Lock()
	defer db.mu.Unlock()

	_ = db.db.View(func(txn *badger.Txn) error {
		fmt.Println("printing all keys")
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				fmt.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *badgerDB) Exists(key []byte) (bool, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	var exists bool
	err := db.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			exists = false
		} else if err == badger.ErrEmptyKey {
			exists = false
		} else if err != nil {
			return err
		} else {
			exists = true
		}
		return nil
	})
	return exists, err
}

func (db *badgerDB) Keys(pref []byte) ([][]byte, error) {

	db.mu.Lock()
	defer db.mu.Unlock()

	var keys [][]byte
	err := db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(pref); it.ValidForPrefix(pref); it.Next() {
			item := it.Item()
			k := item.Key()
			keys = append(keys, k)
		}
		return nil
	})
	return keys, err
}
