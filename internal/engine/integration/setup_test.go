//go:build pglive

// package integration_test contains full engine integration tests
package integration_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/conv"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/internal/sql/registry"
)

func TestMain(m *testing.M) {
	// pg.UseLogger(log.NewStdOut(log.InfoLevel)) // uncomment for debugging
	m.Run()
}

// setup sets up the global context and registry for the tests
func setup(t *testing.T) (global *execution.GlobalContext, reg *registry.Registry, db *pg.DB, err error) {
	ctx := context.Background()

	cfg := &pg.DBConfig{
		PoolConfig: pg.PoolConfig{
			ConnConfig: pg.ConnConfig{
				Host:   "127.0.0.1",
				Port:   "5432",
				User:   "kwild",
				Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
				DBName: "kwil_test_db",
			},
			MaxConns: 11,
		},
		SchemaFilter: func(s string) bool {
			return strings.Contains(s, "ds_")
		},
	}
	db, err = pg.NewDB(ctx, cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	db.AutoCommit(true)        // init and loading writes out-of-session
	defer db.AutoCommit(false) // tests use sessions

	t.Cleanup(func() {
		p := db.Pool()
		_, err := p.Execute(ctx, `DO $$
		DECLARE
			sn text;
		BEGIN
			FOR sn IN SELECT schema_name FROM information_schema.schemata WHERE schema_name LIKE 'ds_%'
			LOOP
				EXECUTE 'DROP SCHEMA ' || quote_ident(sn) || ' CASCADE';
			END LOOP;
		END $$;`)
		if err != nil {
			t.Error(err)
		}
		_, err = p.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_internal CASCADE`)
		if err != nil {
			t.Error(err)
		}

		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}
	})

	reg, err = registry.New(ctx, db, registry.WithLogger(log.NewStdOut(log.InfoLevel)))
	if err != nil {
		return nil, nil, nil, err
	}

	t.Cleanup(func() {
		skey := randomBytes(32)
		err := reg.Begin(ctx, skey)
		if err != nil {
			t.Fatal(err)
		}

		dbids, err := reg.List(ctx)
		if err != nil {
			t.Error(err)
		} else {
			// t.Logf("Removing DBIDs: %v", dbids)
			for _, dbid := range dbids {
				err = reg.Delete(ctx, dbid)
				if err != nil {
					t.Error(err)
				}
			}
		}

		id, err := reg.Precommit(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(id) == 0 {
			t.Error("expected a non-empty commit id")
		}

		err = reg.Commit(ctx, skey)
		if err != nil {
			t.Error(err)
		}

		err = reg.Close(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})

	global, err = execution.NewGlobalContext(ctx, reg, map[string]execution.ExtensionInitializer{
		"math": (&mathInitializer{}).initialize,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return global, reg, db, nil
}

func randomBytes(l int) []byte {
	b := make([]byte, l)
	_, _ = rand.Read(b)
	return b
}

// mocks a namespace initializer
type mathInitializer struct {
	vals map[string]string
}

func (m *mathInitializer) initialize(_ *execution.DeploymentContext, mp map[string]string) (execution.ExtensionNamespace, error) {
	m.vals = mp

	_, ok := m.vals["fail"]
	if ok {
		return nil, fmt.Errorf("mock extension failed to initialize")
	}

	return &mathExt{}, nil
}

type mathExt struct{}

var _ execution.ExtensionNamespace = &mathExt{}

func (m *mathExt) Call(caller *execution.ProcedureContext, method string, inputs []any) ([]any, error) {
	if method != "add" {
		return nil, fmt.Errorf("unknown method: %s", method)
	}

	if len(inputs) != 2 {
		return nil, fmt.Errorf("expected 2 inputs, got %d", len(inputs))
	}

	// The extension needs to tolerate any compatible input type.

	a, err := conv.Int(inputs[0])
	if err != nil {
		return nil, fmt.Errorf("expected int64, got %T (%w)", inputs[0], err)
	}

	b, err := conv.Int(inputs[1])
	if err != nil {
		return nil, fmt.Errorf("expected int64, got %T (%w)", inputs[1], err)
	}

	return []any{a + b}, nil
}
