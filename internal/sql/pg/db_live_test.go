//go:build pglive

package pg

import (
	"context"
	"strings"
	"testing"
)

const (
	pingStmt = `-- ping`
)

func TestMain(m *testing.M) {
	// UseLogger(log.NewStdOut(log.DebugLevel))
	m.Run()
}

func TestNestedTx(t *testing.T) {
	ctx := context.Background()
	cfg := &DBConfig{
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

	// success on outer tx even though failure in a savepoint
	err = tx.Commit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: enure updates in other non-failed savepoints take
}
