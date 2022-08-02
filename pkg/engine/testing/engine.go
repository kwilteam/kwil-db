package testing

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/engine"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
)

func NewTestEngine(ctx context.Context, opts ...engine.EngineOpt) (*engine.Engine, func() error, error) {
	opener := newTestDBOpener()

	opts = append(opts, engine.WithOpener(opener))

	e, err := engine.Open(ctx,
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

func (t *testDbOpener) Open(name, path string, l log.Logger) (engine.Datastore, error) {
	ds, teardown, err := sqlTesting.OpenTestDB(name)
	if err != nil {
		return nil, err
	}

	t.teardowns = append(t.teardowns, teardown)

	return ds, nil
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

type datastoreAdapter struct {
	sqlTesting.TestSqliteClient
}

func (d *datastoreAdapter) Prepare(query string) (engine.Statement, error) {
	return d.TestSqliteClient.Prepare(query)
}

func (d *datastoreAdapter) Savepoint() (engine.Savepoint, error) {
	return d.TestSqliteClient.Savepoint()
}
