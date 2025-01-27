package specifications

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	namespace = "test"

	createNamespace = `CREATE NAMESPACE IF NOT EXISTS ` + namespace + `;`

	dbSchema = `{test}CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY,
		name TEXT NOT NULL,
		age INT NOT NULL
	);

	{test}CREATE TABLE IF NOT EXISTS posts (
		id INT PRIMARY KEY,
		owner_id INT NOT NULL REFERENCES users(id),
		content TEXT,
		created_at INT
	);
	
	{test}CREATE ACTION create_user($id int, $name text, $age int) public {
		INSERT INTO users (id, name, age) 
		VALUES ($id, $name, $age);
	};

	{test}CREATE ACTION list_users() public {
		SELECT * FROM users;
	};
	`
)

func CreateNamespaceSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl, expectFailure bool) {
	txHash, err := execute.ExecuteSQL(ctx, createNamespace, nil)
	require.NoError(t, err)

	if expectFailure {
		expectTxFail(t, execute, ctx, txHash, defaultTxQueryTimeout)()
		return
	}

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func CreateSchemaSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	txHash, err := execute.ExecuteSQL(ctx, dbSchema, nil)
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

type User struct {
	Id   int
	Name string
	Age  int
}

func AddUserSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl, user *User) {
	// add a user
	createCmd := fmt.Sprintf("{%s}INSERT INTO users (id, name, age) VALUES (%d, '%s', %d);", namespace, user.Id, user.Name, user.Age)
	txHash, err := execute.ExecuteSQL(ctx, createCmd, nil)
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func ListUsersSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl, expectFailure bool, numUsers int) {
	res, err := execute.Query(ctx, fmt.Sprintf("{%s}SELECT * FROM users;", namespace), nil)
	if expectFailure {
		require.Error(t, err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, res)
	fmt.Println(res)
	require.Equal(t, numUsers, len(res.Values))
}

func ListUsersEventuallySpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl, expectFailure bool, numUsers int) {
	require.Eventually(t, func() bool {
		res, err := execute.Query(ctx, fmt.Sprintf("{%s}SELECT * FROM users;", namespace), nil)
		return expectFailure && err != nil || res != nil && len(res.Values) == numUsers
	}, 2*time.Minute, 1*time.Second)
}

func AddUserActionSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl, user *User) {
	// add a user
	txHash, err := execute.Execute(ctx, namespace, "create_user", [][]any{{user.Id, user.Name, user.Age}})
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}
