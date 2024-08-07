package specifications

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExecuteContextualVarsSpecification(ctx context.Context, t *testing.T, execute ProcedureDSL) {
	db := SchemaLoader.Load(t, ContextualVarsDB)

	res, err := execute.DeployDatabase(ctx, db)
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	// test procedures
	t.Log("Testing contextual vars procedures")
	testCtxVars(ctx, t, execute, execute.DBID(db.Name), "proc")

	// delete
	res, err = execute.Execute(ctx, execute.DBID(db.Name), "delete_all", []any{})
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	// test actions
	t.Log("Testing contextual vars actions")
	testCtxVars(ctx, t, execute, execute.DBID(db.Name), "act")
}

func testCtxVars(ctx context.Context, t *testing.T, execute ProcedureDSL, dbid, prefix string) {
	res, err := execute.Execute(ctx, dbid, prefix+"_store_vars", []any{})
	require.NoError(t, err)
	ExpectTxSuccess(t, execute, ctx, res)

	results, err := execute.Call(ctx, dbid, "get_stored", []any{})
	require.NoError(t, err)

	count := 0
	for results.Next() {
		count++
		rec := results.Record()
		require.NotNil(t, rec)
		require.Len(t, rec, 6)

		// check caller
		ident, err := execute.Identifier()
		require.NoError(t, err)
		assert.Equal(t, ident, rec["caller"])

		expectedSigner := base64.StdEncoding.EncodeToString(execute.Signer())
		// signer
		assert.Equal(t, expectedSigner, rec["signer"])

		// txid
		expectedRes := hex.EncodeToString(res)
		assert.Equal(t, expectedRes, rec["txid"])

		// height.
		// We don't know the exact height, but it should be greater than 0
		height := rec["height"]
		if height.(int64) <= 0 {
			t.Errorf("height should be greater than 0")
		}

		// block_timestamp
		// We don't know the exact timestamp, but it should be greater than 1722439321
		// (the time I am writing this test), and less than the current time.
		blockTimestamp := rec["block_timestamp"]
		if blockTimestamp.(int64) <= 1722439321 {
			t.Errorf("block_timestamp should be greater than 1722439321")
		}

		// since our test node is acting honestly
		if blockTimestamp.(int64) >= time.Now().Unix()+100 {
			t.Errorf("block_timestamp should be less than the current time")
		}

		// authenticator
		authen := rec["authenticator"]
		assert.Equal(t, auth.EthPersonalSignAuth, authen)
	}
	require.Equal(t, 1, count)
}
