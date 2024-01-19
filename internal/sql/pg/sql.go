package pg

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	// Tables with no PRIMARY KEY or UNIQUE index will fail to update or delete
	// when there is an active publication and replication slot unless the
	// table's "replication identity" is explicitly set to "full". We ensure
	// that is the case by creating an event trigger to perform the ALTER TABLE
	// command whenever a DDL command with the "CREATE TABLE" tag is processed
	// for a table with neither a primary key or unique index. These are the
	// embedded plpgsql functions below.

	//go:embed trigger_repl1.sql
	sqlFuncReplIfNeeded string

	//go:embed trigger_repl2.sql
	sqlFuncReplIfNeeded2 string // I'm still deciding which to use

	sqlCreateFuncReplIdentExists = `SELECT EXISTS (
		SELECT 1 FROM pg_proc 
		WHERE proname = 'set_replica_identity_full'
	);` // checks if the repl trigger created in sqlFuncReplIfNeeded exists

	sqlCreateEvtTriggerReplIdentExists = `SELECT EXISTS (
		SELECT 1 FROM pg_event_trigger 
		WHERE evtname = 'trg_set_replica_identity_full'
	);`

	sqlCreateEvtTriggerReplIdent = `CREATE EVENT TRIGGER trg_set_replica_identity_full ON ddl_command_end
		WHEN TAG IN ('CREATE TABLE')
		EXECUTE FUNCTION set_replica_identity_full();`
	// TIP for node reset/cleanup: DROP EVENT TRIGGER IF EXISTS trg_set_replica_identity_full;

	// TODO: I think we might advise or just support postgres "superuser" access
	// in which case we can do ALL of the dba init tasks, including:
	//
	//  -- <while connected to system 'postgres' database>
	//  CREATE USER kwild WITH SUPERUSER REPLICATION;
	//  CREATE DATABASE kwild OWNER kwild;
	//  -- <reconnect, to the new 'kwild' database>
	//  CREATE PUBLICATION kwild_repl FOR ALL TABLES;
)

const (
	internalSchemaName = "kwild_internal"

	sentryTableName      = `sentry`
	sentryTableNameFull  = internalSchemaName + "." + sentryTableName
	sqlCreateSentryTable = `CREATE TABLE IF NOT EXISTS ` + sentryTableNameFull + ` (seq INT8);`

	sqlInsertSentryRow = `INSERT INTO ` + sentryTableNameFull + ` (seq) VALUES ($1);`
	sqlSelectSentrySeq = `SELECT seq FROM ` + sentryTableNameFull
	sqlUpdateSentrySeq = `UPDATE ` + sentryTableNameFull + ` SET seq = $1;`

	sqlCreateSchemaTemplate = `CREATE SCHEMA IF NOT EXISTS %s;`
	sqlSchemaExists         = `SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name = $1;`

	sqlSchemaTableExists = `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE  table_schema = $1
		AND    table_name   = $2
	);`
	sqlTableExists = `SELECT to_regclass($1);`
)

func tableExists(ctx context.Context, schema, table string, conn *pgx.Conn) (bool, error) {
	rows, _ := conn.Query(ctx, sqlSchemaTableExists, schema, table)
	return pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
}

func ensureTriggerReplIdentity(ctx context.Context, conn *pgx.Conn) error {
	// First crate the function if needed.
	_, err := conn.Exec(ctx, sqlFuncReplIfNeeded)
	if err != nil {
		return err
	}

	// Create the trigger for the function if needed.
	rows, _ := conn.Query(ctx, sqlCreateEvtTriggerReplIdentExists)
	triggerExists, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return err
	}
	if triggerExists {
		return nil
	}
	_, err = conn.Exec(ctx, sqlCreateEvtTriggerReplIdent)
	return err
}

func ensureSentryTable(ctx context.Context, conn *pgx.Conn) error {
	exists, err := tableExists(ctx, internalSchemaName, sentryTableName, conn)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	createStmt := fmt.Sprintf(sqlCreateSchemaTemplate, internalSchemaName)
	_, err = conn.Exec(ctx, createStmt)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, sqlCreateSentryTable)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, sqlInsertSentryRow, 0)
	return err
}

func incrementSeq(ctx context.Context, tx pgx.Tx) (int64, error) {
	var seq int64
	if err := tx.QueryRow(ctx, sqlSelectSentrySeq).Scan(&seq); err != nil {
		return 0, fmt.Errorf("sentry seq scan failed: %w", err)
	}
	seq++
	if res, err := tx.Exec(ctx, sqlUpdateSentrySeq, seq); err != nil {
		return 0, fmt.Errorf("sentry seq update failed: %w", err)
	} else if n := res.RowsAffected(); n != 1 {
		return 0, fmt.Errorf("sentry seq update affected %d rows, not 1", n)
	}
	return seq, nil
	// NOTE: I'd like the above to be a single statement with `RETURNING seq`,
	// but we can't get that value with Exec. At least we have a mutex locked.
}
