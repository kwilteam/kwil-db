package specifications

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type ProcedureDSL interface {
	DatabaseDeployDsl
	ExecuteActionsDsl
	InfoDsl
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

	ex := &executor{
		t:    t,
		dsl:  caller,
		dbid: dbid,
	}

	// create user
	name := "satoshi"
	age := int64(42)

	ex.Execute(ctx, "create_user", []any{name, age})

	// get user
	res, err := ex.Call(ctx, "get_user", []any{name})
	require.NoError(t, err, "error calling get_user action")

	require.Len(t, res, 1)

	user := res[0]

	// the cli returns all numbers as strings
	var ok bool
	age, ok = user["age"].(int64)
	require.Truef(t, ok, "expected an int64, got a %T", user["age"])
	require.Equal(t, age, int64(42))

	// get owned users, returns a table
	res, err = ex.Call(ctx, "get_users_by_age", []any{42})
	require.NoError(t, err, "error calling get_users_by_age action")

	require.Len(t, res, 1)

	user = res[0]

	name, ok = user["name"].(string)
	require.Truef(t, ok, "expected a string, got a %T", user["name"])
	require.Equal(t, name, "satoshi")

	testPosts(ctx, t, ex)
}

type executor struct {
	t    *testing.T
	dsl  ProcedureDSL
	dbid string
}

func (e *executor) Execute(ctx context.Context, actionName string, actionInputs []any, expectFail ...bool) {
	res, err := e.dsl.Execute(ctx, e.dbid, actionName, actionInputs)
	require.NoError(e.t, err, "error executing action")

	if len(expectFail) > 0 && expectFail[0] {
		expectTxFail(e.t, e.dsl, ctx, res, defaultTxQueryTimeout)()
	} else {
		expectTxSuccess(e.t, e.dsl, ctx, res, defaultTxQueryTimeout)()
	}
}

func (e *executor) Call(ctx context.Context, action string, inputs []any) ([]map[string]any, error) {
	res, err := e.dsl.Call(ctx, e.dbid, action, inputs)
	if err != nil {
		return nil, err
	}

	return res.Export(), nil
}

// executeProcedureReturnNextSpecification tests that procedures properly handle
// RETURN NEXT semantics. This is kept unexported because creating a user in the database
// is a pre-requisite, which is done in the exported ExecuteProcedureSpecification.
// This test uses the `create_procedure` and `get_recent_posts_by_size` procedures from users.kf.
func testPosts(ctx context.Context, t *testing.T, caller *executor) {
	// we will makie 5 posts, with 3 of them having more than 100 characters
	posts := []string{
		"short_post_1",
		"long1_" + strings.Repeat("a", 100),
		"long2_" + strings.Repeat("b", 100),
		"long3_" + strings.Repeat("c", 100),
		"short_post_2",
	}

	// create posts
	for _, post := range posts {
		caller.Execute(ctx, "create_post", []any{post})
	}

	// get recent posts
	res, err := caller.Call(ctx, "get_recent_posts", []any{"satoshi"})
	require.NoError(t, err, "error calling get_recent_posts_by_size action")

	require.Len(t, res, 5)

	// get recent posts by size, will limit to return 2
	res, err = caller.Call(ctx, "get_recent_posts_by_size", []any{"satoshi", 100, 2})
	require.NoError(t, err, "error calling get_recent_posts_by_size action")

	require.Len(t, res, 2)

	// check that the posts are ordered by size, and the
	// latest posts are returned first
	require.Equal(t, res[0]["content"], posts[3])
	require.Equal(t, res[1]["content"], posts[2])

	// reverse latest posts
	res, err = caller.Call(ctx, "reverse_latest_posts", []any{"satoshi", 5})
	require.NoError(t, err, "error calling reverse_latest_posts action")

	require.Len(t, res, 1)

	content, ok := res[0]["content"].([]interface{})
	require.True(t, ok)

	// check that the posts are reversed. since the query is reversing
	// latest posts, it is expected that the posts are in the order
	// they were created
	for i, c := range content {
		con, ok := c.(string)
		require.Truef(t, ok, "wanted a string element, got a %T", c)

		require.Equal(t, con, posts[i])
	}
}
