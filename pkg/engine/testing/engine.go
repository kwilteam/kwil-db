package testing

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql"

	"github.com/kwilteam/kwil-db/pkg/engine"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
)

func NewTestEngine(ctx context.Context, ec engine.CommitRegister, opts ...engine.EngineOpt) (*engine.Engine, func() error, error) {
	opener := newTestDBOpener()

	e, err := engine.Open(ctx, opener, ec,
		opts...,
	)
	if err != nil {
		return nil, nil, err
	}

	return e, opener.Teardown, nil
}

func newTestDBOpener() *testDbOpener {
	return &testDbOpener{
		teardowns: make([]func() error, 0),
	}
}

// testDbOpener creates real sqlite databases that can be used for testing
// it also keeps track of the teardown functions so that they can be called
// after the test is complete
type testDbOpener struct {
	teardowns []func() error
}

// implements sql.Opener
func (t *testDbOpener) Open(name string, _ log.Logger) (sql.Database, error) {
	db, td, err := sqlTesting.OpenTestDB(name)
	if err != nil {
		return nil, err
	}

	t.teardowns = append(t.teardowns, td)

	return db, nil
}

func (t *testDbOpener) Teardown() error {
	var errs []error
	for _, teardown := range t.teardowns {
		err := teardown()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
