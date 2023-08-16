package testing

import (
	"errors"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/pkg/kv"
	"github.com/kwilteam/kwil-db/pkg/kv/atomic"
	"github.com/kwilteam/kwil-db/pkg/kv/badger"
)

type TestKVFlag uint8

const (
	TestKVFlagInMemory TestKVFlag = iota
	TestKVFlagBadger
)

const defaultPath = "./tmp/"

// OpenTestKv opens a new test kv store
// It returns a teardown function.  If a teardown
// function is not necessary, it does nothing
func OpenTestKv(name string, flag TestKVFlag) (*atomic.AtomicKV, func() error, error) {
	directory := fmt.Sprintf("%s%s", defaultPath, name)

	switch flag {
	case TestKVFlagInMemory:
		fn := func() error {
			return nil
		}

		db, err := atomic.NewAtomicKV(newMemoryKV())
		return db, fn, err
	case TestKVFlagBadger:
		badgerDB, err := badger.NewBadgerDB(directory)
		if err != nil {
			return nil, nil, err
		}

		fn := func() error {
			return errors.Join(
				badgerDB.Close(),
				os.RemoveAll(directory),
			)
		}

		db, err := atomic.NewAtomicKV(badgerDB)
		return db, fn, err
	default:
		return nil, nil, fmt.Errorf("unknown flag: %d", flag)
	}
}

func newMemoryKV() *MemoryKV {
	return &MemoryKV{
		values: make(map[string][]byte),
	}
}

type MemoryKV struct {
	values map[string][]byte
}

func (m *MemoryKV) BeginTransaction() kv.Transaction {

	return &MemoryTransaction{
		kv:             m,
		currentTx:      make(map[string][]byte),
		currentDeletes: make(map[string]struct{}),
	}
}

func (m *MemoryKV) Delete(key []byte) error {
	_, ok := m.values[string(key)]
	if !ok {
		return kv.ErrKeyNotFound
	}

	delete(m.values, string(key))

	return nil
}

func (m *MemoryKV) Get(key []byte) ([]byte, error) {
	val, ok := m.values[string(key)]
	if !ok {
		return nil, kv.ErrKeyNotFound
	}

	return val, nil
}

func (m *MemoryKV) Set(key []byte, value []byte) error {
	m.values[string(key)] = value

	return nil
}

type MemoryTransaction struct {
	currentTx      map[string][]byte
	currentDeletes map[string]struct{}
	kv             *MemoryKV
}

func (m *MemoryTransaction) Commit() error {
	for k, v := range m.currentTx {
		m.kv.values[k] = v
	}

	for k := range m.currentDeletes {
		delete(m.kv.values, k)
	}

	m.currentTx = nil

	return nil
}

func (m *MemoryTransaction) Delete(key []byte) error {
	_, err := m.Get(key)
	if err != nil {
		return err
	}

	m.currentDeletes[string(key)] = struct{}{}

	return nil
}

func (m *MemoryTransaction) Discard() {
	m.currentTx = nil
	m.currentDeletes = nil
}

func (m *MemoryTransaction) Get(key []byte) ([]byte, error) {
	val, ok := m.currentTx[string(key)]
	if ok {
		return val, nil
	}

	val, ok = m.kv.values[string(key)]
	if ok {
		return val, nil
	}

	return nil, kv.ErrKeyNotFound
}

func (m *MemoryTransaction) Set(key []byte, value []byte) error {
	m.currentTx[string(key)] = value

	return nil
}
