package specifications

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/kwilteam/kwil-db/test/driver"
	"github.com/stretchr/testify/require"
)

type DatabaseSchemaLoader interface {
	Load(t *testing.T, targetSchema *testSchema) *types.Schema
	LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *types.Schema
}

type FileDatabaseSchemaLoader struct {
	Modifier func(db *types.Schema)
}

func (l *FileDatabaseSchemaLoader) Load(t *testing.T, targetSchema *testSchema) *types.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	parseResult, err := parse.Parse(d)
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}

	l.Modifier(parseResult)
	return parseResult
}

func (l *FileDatabaseSchemaLoader) LoadWithoutValidation(t *testing.T, targetSchema *testSchema) *types.Schema {
	t.Helper()

	d, err := os.ReadFile(targetSchema.GetFilePath())
	if err != nil {
		t.Fatal("cannot open database schema file", err)
	}

	db, err := parse.ParseSchemaWithoutValidation(d)
	if err != nil {
		t.Fatal("cannot parse database schema", err)
	}
	// ignore parser validation error

	l.Modifier(db.Schema)

	return db.Schema
}

func ExpectTxSuccess(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte) {
	expectTxSuccess(t, spec, ctx, txHash, defaultTxQueryTimeout)()
}

func expectTxSuccess(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte, waitFor time.Duration) func() {
	return func() {
		var status strings.Builder
		require.Eventually(t, func() bool {
			// prevent appending to the prior invocation(s)
			status.Reset()
			if err := spec.TxSuccess(ctx, txHash); err == nil {
				return true
				// Consider failing faster for unexpected errors:
				// } else if !errors.Is(err, driver.ErrTxNotConfirmed) {
				// 	t.Fatal(err)
				// 	return false
			} else {
				status.WriteString(err.Error())
				return false
			}
		}, waitFor, time.Millisecond*300, "tx failed: %s", status.String())
	}
}

func ExpectTxfail(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte) {
	expectTxFail(t, spec, ctx, txHash, defaultTxQueryTimeout)()
}

func expectTxFail(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte, waitFor time.Duration) func() {
	return func() {
		var status strings.Builder
		require.Eventually(t, func() bool {
			// prevent appending to the prior invocation(s)
			status.Reset()
			if err := spec.TxSuccess(ctx, txHash); err == nil {
				status.WriteString("success")
				return false
			} else {
				status.WriteString(err.Error())
				// NOTE: ErrTxNotConfirmed is not considered a failure, should retry
				return !errors.Is(err, driver.ErrTxNotConfirmed)
			}
		}, waitFor, time.Second*1, "tx should fail - status: %v, hash %x", status.String(), txHash)
	}
}
