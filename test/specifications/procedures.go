package specifications

import (
	"context"
	"testing"

	"github.com/cstockton/go-conv"
	"github.com/stretchr/testify/require"
)

type ProcedureDSL interface {
	DatabaseDeployDsl
	ExecuteActionsDsl
}

// ExecuteProcedureSpecification tests that procedures for a specific schema
// can be generated and executed. It handles deployment of the schema, as well
// as the calling of procedures.
func ExecuteProcedureSpecification(ctx context.Context, t *testing.T, caller ProcedureDSL) {
	schema := SchemaLoader.Load(t, userDB)

	// deploy
	txHash, err := caller.DeployDatabase(ctx, schema)
	require.NoError(t, err, "error deploying schema")

	expectTxSuccess(t, caller, ctx, txHash, defaultTxQueryTimeout)()

	dbid := caller.DBID(schema.Name)

	// create user
	name := "satoshi"
	age := int64(42)

	txHash, err = caller.Execute(ctx, dbid, "create_user", []any{name, age})
	require.NoError(t, err, "error executing create_user action")

	expectTxSuccess(t, caller, ctx, txHash, defaultTxQueryTimeout)()

	// get user
	res, err := caller.Call(ctx, dbid, "get_user", []any{name})
	require.NoError(t, err, "error calling get_user action")

	records := res.Export()
	require.Len(t, records, 1)

	user := records[0]

	// we use conv here because the cli returns all numbers as strings
	age, err = conv.Int64(user["age"])
	require.NoError(t, err)
	require.Equal(t, age, int64(42))

	// get owned users, returns a table
	res, err = caller.Call(ctx, dbid, "get_users_by_age", []any{42})
	require.NoError(t, err, "error calling get_users_by_age action")

	records = res.Export()
	require.Len(t, records, 1)

	user = records[0]

	name, ok := user["name"].(string)
	require.True(t, ok)
	require.Equal(t, name, "satoshi")
	_, ok = user["address"].(string)
	require.True(t, ok)

	// create post
	content := "hello world"

	txHash, err = caller.Execute(ctx, dbid, "create_post", []any{content})
	require.NoError(t, err, "error executing create_post action")

	expectTxSuccess(t, caller, ctx, txHash, defaultTxQueryTimeout)()
}
