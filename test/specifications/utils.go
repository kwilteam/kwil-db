package specifications

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/kwilteam/kwil-db/test/driver"
	"github.com/stretchr/testify/assert"
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
	expectTxSuccess(t, spec, ctx, txHash, defaultTxQueryTimeout)
}

func expectTxSuccess(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte, waitFor time.Duration) {
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		err := spec.TxSuccess(ctx, txHash)
		if err == nil {
			return // stop checking
		}

		require.ErrorIs(collect, err, driver.ErrTxNotConfirmed) // fail fast if unexpected error

		collect.Errorf("not confirmed") // keep checking
	}, waitFor, time.Millisecond*300, "tx did not succeed")
}

func ExpectTxFail(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte) {
	expectTxFail(t, spec, ctx, txHash, defaultTxQueryTimeout)
}

// expectTxFail should fail if spec.TxSuccess returns an error that is NOT of
// type driver.ErrTxNotConfirmed. It will keep checking while the error IS
// driver.ErrTxNotConfirmed. If spec.TxSuccess return without error, this should
// also fail the test (TODO: it keeps checking until waitFor timeout!)
func expectTxFail(t *testing.T, spec TxQueryDsl, ctx context.Context, txHash []byte, waitFor time.Duration) {
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		err := spec.TxSuccess(ctx, txHash)
		require.Error(collect, err, "transaction succeeded") // fail fast with require if it executed without error

		assert.NotErrorIs(collect, err, driver.ErrTxNotConfirmed) // tick again if not confirmed

		// otherwise it failed (yay) and we're done checking (raise no errors in this tick)
	}, waitFor, time.Second*1, "tx should have failed")
}
