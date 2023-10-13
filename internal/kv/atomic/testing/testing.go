package testing

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/kv/atomic"
	badgerTesting "github.com/kwilteam/kwil-db/internal/kv/badger/testing"
	kvTesting "github.com/kwilteam/kwil-db/internal/kv/testing"
)

type TestKVFlag uint8

const (
	TestKVFlagInMemory TestKVFlag = iota
	TestKVFlagBadger
)

// OpenTestKv opens a new test kv store
// It returns a teardown function.  If a teardown
// function is not necessary, it does nothing
func OpenTestKv(ctx context.Context, name string, flag TestKVFlag) (*atomic.AtomicKV, func() error, error) {

	switch flag {
	case TestKVFlagInMemory:
		fn := func() error {
			return nil
		}

		db, err := atomic.NewAtomicKV(kvTesting.NewMemoryKV())
		return db, fn, err
	case TestKVFlagBadger:
		badgerDB, td, err := badgerTesting.NewTestBadgerDB(ctx, name, nil)
		if err != nil {
			return nil, nil, err
		}

		db, err := atomic.NewAtomicKV(badgerDB)
		return db, td, err
	default:
		return nil, nil, fmt.Errorf("unknown flag: %d", flag)
	}
}
