//go:build pglive

package node

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/trufnetwork/kwil-db/config"
	"github.com/trufnetwork/kwil-db/core/log"
	"github.com/trufnetwork/kwil-db/node/pg"
)

// restoreTestDB creates an empty database for one test and drops it after.
func restoreTestDB(t *testing.T, name string) config.DBConfig {
	t.Helper()
	ctx := context.Background()
	db := config.DBConfig{Host: "127.0.0.1", Port: "5432", User: "kwild", Pass: "kwild", DBName: name}

	admin, err := pgx.Connect(ctx, fmt.Sprintf("host=%s port=%s user=%s password=%s database=postgres sslmode=disable",
		db.Host, db.Port, db.User, db.Pass))
	require.NoError(t, err)
	drop := func() {
		_, err := admin.Exec(ctx, `DROP DATABASE IF EXISTS `+name+` WITH (FORCE)`)
		require.NoError(t, err)
	}
	drop()
	_, err = admin.Exec(ctx, `CREATE DATABASE `+name+` OWNER kwild`)
	require.NoError(t, err)
	t.Cleanup(func() {
		drop()
		admin.Close(ctx)
	})
	return db
}

func tablesIn(t *testing.T, db config.DBConfig) []string {
	t.Helper()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Pass, db.DBName))
	require.NoError(t, err)
	defer conn.Close(ctx)
	rows, _ := conn.Query(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = 'restored' ORDER BY 1`)
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)
	return names
}

func restore(t *testing.T, db config.DBConfig, dump string) error {
	sum := sha256.Sum256([]byte(dump))
	return RestoreDB(context.Background(), strings.NewReader(dump), db, sum[:], log.DiscardLogger)
}

// openAsKwild opens db the way kwild does at every start, then closes it. That
// installs kwild's event trigger, which sets each table created after it to
// REPLICA IDENTITY FULL.
func openAsKwild(t *testing.T, db config.DBConfig) {
	t.Helper()
	kwild, err := pg.NewDB(context.Background(), &pg.DBConfig{PoolConfig: pg.PoolConfig{ConnConfig: *pgConnConfig(db), MaxConns: 2}})
	require.NoError(t, err)
	require.NoError(t, kwild.Close())
}

// replicaIdentities maps each table in schema to its replica identity as
// pg_class records it: d for DEFAULT, f for FULL.
func replicaIdentities(t *testing.T, db config.DBConfig, schema string) map[string]string {
	t.Helper()
	rows, err := dbConn(t, db).Query(context.Background(), `SELECT c.relname, c.relreplident::text
		FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relkind = 'r'`, schema)
	require.NoError(t, err)
	identities := map[string]string{}
	for rows.Next() {
		var table, identity string
		require.NoError(t, rows.Scan(&table, &identity))
		identities[table] = identity
	}
	require.NoError(t, rows.Err())
	return identities
}

func TestRestoreDBStopsAtTheFirstFailedStatement(t *testing.T) {
	db := restoreTestDB(t, "kwil_test_restore_error")
	dump := "CREATE SCHEMA restored;\n" +
		"CREATE TABLE restored.before (v int8);\n" +
		"COPY restored.before (v) FROM stdin;\n" +
		"1\n" +
		"not a number\n" +
		"\\.\n" +
		"CREATE TABLE restored.after (v int8);\n" +
		// More than a pipe buffer after the failure, so psql exits while the
		// dump is still being written and the write sees a broken pipe.
		strings.Repeat("SELECT 1;\n", 200_000)

	err := restore(t, db, dump)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid input syntax for type bigint")
	require.Equal(t, []string{"before"}, tablesIn(t, db), "psql kept going after the failed COPY")
}

func TestRestoreDBAcceptsRestrictLines(t *testing.T) {
	db := restoreTestDB(t, "kwil_test_restore_restrict")
	// The shape pg_dump 16.10 and later give a sanitized snapshot.
	dump := "\\restrict key1\n" +
		"\\unrestrict key1\n" +
		"\\restrict key1\n" +
		"CREATE SCHEMA restored;\n" +
		"CREATE TABLE restored.blocks (v int8);\n" +
		"COPY restored.blocks (v) FROM stdin;\n" +
		"1\n" +
		"\\.\n" +
		"\\unrestrict key1\n"

	require.NoError(t, restore(t, db, dump))
	require.Equal(t, []string{"blocks"}, tablesIn(t, db))
}

func TestRestoreDBRejectsAWrongHash(t *testing.T) {
	db := restoreTestDB(t, "kwil_test_restore_hash")
	dump := "CREATE SCHEMA restored;\n"
	wrong := sha256.Sum256([]byte("something else"))

	err := RestoreDB(context.Background(), strings.NewReader(dump), db, wrong[:], log.DiscardLogger)
	require.ErrorContains(t, err, "invalid snapshot hash")
}

func TestRestoreDBKeepsRestrictTextInsideAFunction(t *testing.T) {
	db := restoreTestDB(t, "kwil_test_restore_function")
	dump := "\\restrict wrapkey\n" +
		"CREATE SCHEMA restored;\n" +
		fnWithRestrictText +
		"\\unrestrict wrapkey\n"

	require.NoError(t, restore(t, db, dump))

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Pass, db.DBName))
	require.NoError(t, err)
	defer conn.Close(ctx)
	var got string
	require.NoError(t, conn.QueryRow(ctx, `SELECT restored.f()`).Scan(&got))
	require.Equal(t, "\n\\restrict key1\n", got)
}

func TestRestoreDBKeepsTheDumpsReplicaIdentities(t *testing.T) {
	db := restoreTestDB(t, "kwil_test_restore_identity")
	openAsKwild(t, db)
	// pg_dump writes REPLICA IDENTITY only for a table whose identity is not
	// DEFAULT.
	dump := "CREATE SCHEMA restored;\n" +
		"CREATE TABLE restored.by_key (id int8 PRIMARY KEY, v text);\n" +
		"CREATE TABLE restored.whole_row (id int8 PRIMARY KEY, v text);\n" +
		"ALTER TABLE ONLY restored.whole_row REPLICA IDENTITY FULL;\n"

	require.NoError(t, restore(t, db, dump))
	require.Equal(t, map[string]string{"by_key": "d", "whole_row": "f"}, replicaIdentities(t, db, "restored"))

	// The trigger is still there for the tables kwild creates afterwards.
	_, err := dbConn(t, db).Exec(context.Background(), `CREATE TABLE restored.later (id int8 PRIMARY KEY)`)
	require.NoError(t, err)
	require.Equal(t, "f", replicaIdentities(t, db, "restored")["later"])
}
