package specifications

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kuneiform/kfparser"
)

type DatabaseIdentifier interface {
	DBID(name string) string
}

type DatabaseExister interface {
	DatabaseExists(ctx context.Context, dbid string) error
}

type DatabaseSchemaLoader interface {
	Load(t *testing.T, targetSchema *testSchema) *transactions.Schema
	LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *transactions.Schema
}

type FileDatabaseSchemaLoader struct {
	Modifier func(db *transactions.Schema)
}

func (l *FileDatabaseSchemaLoader) Load(t *testing.T, targetSchema *testSchema) *transactions.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	astSchema, err := kfparser.Parse(string(d))
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}

	schemaJson, err := astSchema.ToJSON()
	if err != nil {
		t.Fatal("failed to marshal schema: %w", err)
	}

	var db *transactions.Schema
	err = json.Unmarshal(schemaJson, &db)
	if err != nil {
		t.Fatal("failed to unmarshal schema json: %w", err)
	}

	l.Modifier(db)
	return db
}

func (l *FileDatabaseSchemaLoader) LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *transactions.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	astSchema, err := kfparser.Parse(string(d))
	// ignore validation error
	if astSchema == nil {
		t.Fatal("cannot parse database schema", err)
	}

	schemaJson, err := json.Marshal(astSchema)
	if err != nil {
		t.Fatal("failed to marshal schema: %w", err)
	}

	var db *transactions.Schema
	err = json.Unmarshal(schemaJson, &db)
	if err != nil {
		t.Fatal("failed to unmarshal schema json: %w", err)
	}

	l.Modifier(db)

	return db
}

func expectTxSuccess(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte, waitFor time.Duration) func() {
	return func() {
		var status strings.Builder
		require.Eventually(t, func() bool {
			// prevent appending to the prior invocation(s)
			status.Reset()
			if err := spec.TxSuccess(ctx, txHash); err == nil {
				return true
			} else {
				status.WriteString(err.Error())
				return false
			}
		}, waitFor, time.Second*1, "tx failed: %s", status.String())
	}
}

func expectTxFail(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte, waitFor time.Duration) func() {
	return func() {
		var status strings.Builder
		require.Eventually(t, func() bool {
			// prevent appending to the prior invocation(s)
			status.Reset()
			if err := spec.TxSuccess(ctx, txHash); err == nil {
				return false
			} else {
				status.WriteString(err.Error())
				return true
			}
		}, waitFor, time.Second*1, "tx should fail", status.String())
	}
}
