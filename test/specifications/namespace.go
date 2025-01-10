package specifications

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	namespace        = "kwil_test"
	invalidNamespace = "dummy_test"

	createNamespace = `CREATE NAMESPACE IF NOT EXISTS ` + namespace + `;`

	createUsersTable = `CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY,
		name TEXT NOT NULL,
		age INT NOT NULL
	);`

	createPostsTable = `CREATE TABLE IF NOT EXISTS posts (
		id INT PRIMARY KEY,
		owner_id INT NOT NULL REFERENCES users(id),
		content TEXT,
		created_at INT
	);`

	createUserAction = `CREATE ACTION create_user($id, $name , $age) public {
	INSERT INTO users (id, name, age) 
	VALUES ($id, $name, $age);
	};`

	listUsersAction = `CREATE ACTION list_users() public {
		SELECT *
		FROM users;
	}`

	dbSchema = `CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY,
		name TEXT NOT NULL,
		age INT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INT PRIMARY KEY,
		owner_id INT NOT NULL REFERENCES users(id),
		content TEXT,
		created_at INT
	);
	
	CREATE ACTION list_users() public {
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

func CreateTablesSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	invalidCreateCmd := fmt.Sprintf("{%s}%s", invalidNamespace, createUsersTable)
	txHash, err := execute.ExecuteSQL(ctx, invalidCreateCmd, nil)
	require.NoError(t, err)

	expectTxFail(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	createCmd := fmt.Sprintf("{%s}%s", namespace, createUsersTable)
	txHash, err = execute.ExecuteSQL(ctx, createCmd, nil)
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func CreateUserSQLSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	createCmd := fmt.Sprintf("{%s}INSERT INTO users (id, name, age) VALUES (1, 'satoshi', 42);", namespace)
	txHash, err := execute.ExecuteSQL(ctx, createCmd, nil)
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func CreateUserActionSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	txHash, err := execute.Execute(ctx, namespace, createUserAction, nil)
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func ListUsersActionSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	txHash, err := execute.Execute(ctx, namespace, listUsersAction, nil)
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()
}

func CreateSchemaSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	txHash, err := execute.ExecuteSQL(ctx, dbSchema, nil)
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	// create_user action
	txHash, err = execute.Execute(ctx, namespace, createUserAction, nil)
	require.NoError(t, err)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	// add a user
	createCmd := fmt.Sprintf("{%s}INSERT INTO users (id, name, age) VALUES (1, 'satoshi', 42);", namespace)
	txHash, err = execute.ExecuteSQL(ctx, createCmd, nil)

	expectTxSuccess(t, execute, ctx, txHash, defaultTxQueryTimeout)()

	// list_users action
	res, err := execute.Query(ctx, fmt.Sprintf("{%s}SELECT * FROM users;", namespace), nil)
	require.NoError(t, err)
	fmt.Println(res)
	// require.Len(t, res.Rows, 1)

}
