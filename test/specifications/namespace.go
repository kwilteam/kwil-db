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

	createUserAction = `CREATE ACTION create_user($id INT, $name TEXT, $age INT) public {INSERT INTO users (id, name, age) VALUES (id, name, age);}`

	listUsersAction = `CREATE ACTION list_users() public {
		SELECT *
		FROM users;
	}`
)

func CreateNamespaceSpecification(ctx context.Context, t *testing.T, execute ExecuteQueryDsl) {
	txHash, err := execute.ExecuteSQL(ctx, createNamespace, nil)
	require.NoError(t, err)

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
