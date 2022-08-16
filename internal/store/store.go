package store

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

const (
	// Default BadgerDB discardRatio. It represents the discard ratio for the
	// BadgerDB GC.
	//
	// Ref: https://godoc.org/github.com/dgraph-io/badger#DB.RunValueLogGC
	badgerDiscardRatio = 0.5

	// Default BadgerDB GC interval
	badgerGCInterval = 10 * time.Minute
)

type BadgerDB struct {
	db  *badger.DB
	log zerolog.Logger
}

func New(conf *types.Config) (*BadgerDB, error) {
	// create logger
	logger := log.With().Str("module", "store").Int64("chainID", int64(conf.ClientChain.GetChainID())).Logger()

	// create badger db
	opts := badger.DefaultOptions(conf.Storage.Badger.Path)
	opts.SyncWrites = true
	opts.Logger = nil

	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerDB{
		db:  badgerDB,
		log: logger,
	}, nil
}

func (db *BadgerDB) Close() error {
	return db.db.Close()
}

func (db *BadgerDB) Get(key []byte) ([]byte, error) {
	var val []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return types.ErrNotFound
			} else {
				return err
			}
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	return val, err
}

func (db *BadgerDB) Set(key, val []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

func (db *BadgerDB) RunGC() {
	ticker := time.NewTicker(badgerGCInterval)
	for {
		select {
		case <-ticker.C:
			err := db.db.RunValueLogGC(badgerDiscardRatio)
			if err != nil {
				// don't report error when GC didn't result in any cleanup
				if err == badger.ErrNoRewrite {
					db.log.Debug().Err(err).Msg("no cleanup done")
				} else {
					db.log.Error().Err(err).Msg("failed to run GC")
				}
			}

			// Commenting this out since this doesn't currently have context
			/*case <-db.ctx.Done():
			return
			*/
		}
	}
}

func (db *BadgerDB) Delete(key []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (db *BadgerDB) NewTransaction(writeable bool) *badger.Txn {
	return db.db.NewTransaction(writeable)
}

func (db *BadgerDB) DeleteByPrefix(prefix []byte) error {
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
	db.db.View(func(txn *badger.Txn) error {
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

func (db *BadgerDB) GetAllByPrefix(pref []byte) ([][]byte, [][]byte, error) {
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

func (db *BadgerDB) PrintAll() {
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
