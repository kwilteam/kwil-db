package testing

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/pkg/kv/badger"
)

const defaultPath = "./tmp/"

// NewTestBadgerDB returns a new badger db for testing
// it also returns a teardown function, which will remove
// the db and the directory
func NewTestBadgerDB(ctx context.Context, name string, options *badger.Options) (*badger.BadgerDB, func() error, error) {
	directory := fmt.Sprintf("%s%s", defaultPath, name)

	db, err := badger.NewBadgerDB(ctx, directory, nil)
	if err != nil {
		return nil, nil, err
	}

	fn := func() error {
		return errors.Join(
			db.Close(),
			os.RemoveAll(defaultPath),
		)
	}
	return db, fn, err
}
