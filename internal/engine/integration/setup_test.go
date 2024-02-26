//go:build pglive

// package integration_test contains full engine integration tests
package integration_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/extensions/actions"
	"github.com/kwilteam/kwil-db/internal/conv"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// pg.UseLogger(log.NewStdOut(log.InfoLevel)) // uncomment for debugging
	m.Run()
}

// cleanup deletes all schemas and closes the database
func cleanup(t *testing.T, db *pg.DB) {
	db.AutoCommit(true)
	defer db.AutoCommit(false)
	defer db.Close()
	ctx := context.Background()

	_, err := db.Execute(ctx, `DO $$
	DECLARE
		sn text;
	BEGIN
		FOR sn IN SELECT schema_name FROM information_schema.schemata WHERE schema_name LIKE 'ds_%'
		LOOP
			EXECUTE 'DROP SCHEMA ' || quote_ident(sn) || ' CASCADE';
		END LOOP;
	END $$;`)
	require.NoError(t, err)

	_, err = db.Execute(ctx, `DROP SCHEMA IF EXISTS kwild_internal CASCADE`)
	require.NoError(t, err)
}

// setup sets up the global context and registry for the tests
func setup(t *testing.T) (global *execution.GlobalContext, db *pg.DB, err error) {
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
			return strings.Contains(s, pg.DefaultSchemaFilterPrefix)
		},
	}
	db, err = pg.NewDB(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	global, err = execution.NewGlobalContext(ctx, tx, map[string]actions.ExtensionInitializer{
		"math": (&mathInitializer{}).initialize,
	}, &common.Service{
		Logger:           log.NewNoOp().Sugar(),
		ExtensionConfigs: map[string]map[string]string{},
	})
	require.NoError(t, err)

	_, err = tx.Precommit(ctx)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	return global, db, nil
}

// mocks a namespace initializer
type mathInitializer struct {
	vals map[string]string
}

func (m *mathInitializer) initialize(_ *actions.DeploymentContext, _ *common.Service, mp map[string]string) (actions.ExtensionNamespace, error) {
	m.vals = mp

	_, ok := m.vals["fail"]
	if ok {
		return nil, fmt.Errorf("mock extension failed to initialize")
	}

	return &mathExt{}, nil
}

type mathExt struct{}

var _ actions.ExtensionNamespace = &mathExt{}

func (m *mathExt) Call(caller *actions.ProcedureContext, _ *common.App, method string, inputs []any) ([]any, error) {
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
