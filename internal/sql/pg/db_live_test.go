//go:build pglive

package pg

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	// "github.com/kwilteam/kwil-db/internal/conv"
)

func TestMain(m *testing.M) {
	// UseLogger(log.NewStdOut(log.DebugLevel))
	m.Run()
}

const (
	pingStmt = `-- ping`
)

var (
	cfg = &DBConfig{
		PoolConfig: PoolConfig{
			ConnConfig: ConnConfig{
				Host:   "/var/run/postgresql",
				Port:   "",
				User:   "kwil_test_user",
				Pass:   "kwil", // would be ignored if pg_hba.conf set with trust
				DBName: "kwil_test_db",
			},
			MaxConns: 11,
		},
		SchemaFilter: func(s string) bool {
			return strings.Contains(s, "ds_")
		},
	}
)

// TestSelectLiteralType ensures (and demonstrates) that simpler query execution
// modes can effectively handle inline queries like `SELECT $1;` when provided
// and argument that is not a string, which fails in the expanded execution
// modes that try to obtain argument OIDs (postgres data types) via a
// prepare/describe request to the postgres process.
func TestSelectLiteralType(t *testing.T) {
	ctx := context.Background()
	connStr := connString(cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.DBName, false)
	pgCfg, err := pgx.ParseConfig(connStr)
	if err != nil {
		t.Fatal(err)
	}
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement // default
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeExec
	// pgCfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(ctx, pgCfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := conn.Close(ctx); err != nil {
			t.Error(err)
		}
	})

	// var arg any = int64(1)
	// args := []any{arg, arg}
	argMap := map[string]any{
		"a": int64(1),
		"b": int64(1),
	}
	var arg any = pgx.NamedArgs(argMap)
	args := []any{arg}
	stmt := `SELECT @a + @b;`

	// args := []any{arg}
	// stmt := `SELECT $1;`
	// TODO: make more thorough in-line expression test cases

	rows, err := queryImpliedArgTypes(ctx, conn, stmt, args...)
	if err != nil {
		t.Fatal(err)
	}

	// rows, err := conn.Query(ctx, stmt, arg, arg)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// Tip: try the named argument rewriter as such:
	// argMap := map[string]any{
	// 	"a": int64(1),
	// }
	// var arg any = pgx.NamedArgs(argMap)
	//  ^ with `SELECT @a;`
	// (with the same result)

	defer rows.Close()

	for rows.Next() {
		// rows.Values() // []any
		var val any
		err = rows.Scan(&val)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%v (%T)", val, val)
		// valInt, err := conv.Int(val)
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// t.Log(valInt, "(int64)")
	}

	err = rows.Err()
	if err != nil {
		t.Fatal(err)
	}

	// Now with our high level func and mode.
	args2 := append([]any{QueryModeInferredArgTypes}, args...)
	results, err := query(ctx, &cqWrapper{conn}, stmt, args2...)
	if err != nil {
		t.Fatal(err)
	}
	for _, val := range results.Rows[0] {
		t.Logf("%v (%T)", val, val)
	}
}

func TestNestedTx(t *testing.T) {
	ctx := context.Background()

	db, err := NewDB(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Query(ctx, pingStmt)
	if err != nil {
		t.Fatal(err)
	}

	// Start the outer transaction.
	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Query(ctx, pingStmt)
	if err != nil {
		t.Fatal(err)
	}

	// err = tx.Commit(ctx)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// tx.begintx below would then error with "tx is closed" (TODO: test that)

	// Start savepoint 0
	txNested, err := tx.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer txNested.Rollback(ctx)

	// OK query
	_, err = txNested.Query(ctx, pingStmt)
	if err != nil {
		t.Fatal(err)
	}

	// release savepoint 0
	err = txNested.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Start savepoint 1
	txNested2, err := tx.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Query error
	_, err = txNested2.Query(ctx, `SELECT notathing;`)
	if err == nil {
		t.Fatal("should have errored") // expect error
	}
	// if Commit now, should say: ERROR: current transaction is aborted, commands ignored until end of transaction block (SQLSTATE 25P02)

	// rollback savepoint 1 containing the error
	err = txNested2.Rollback(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Now we should be able to keep going.  Make savepoint 3:
	txNested3, err := tx.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer txNested3.Rollback(ctx)

	_, err = txNested3.Query(ctx, pingStmt)
	if err != nil {
		t.Fatal(err)
	}

	err = txNested3.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Commit(ctx)
	if err == nil {
		t.Fatalf("commit should have errored without precommit first")
	}

	id, err := tx.Precommit(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("commit id: %x", id)

	// success on outer tx even though failure in a savepoint
	err = tx.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: enure updates in other non-failed savepoints take
}
