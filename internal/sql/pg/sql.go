package pg

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

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
	sqlFuncReplIfNeeded2 string //nolint:unused
	// (I'm still deciding which to use)

	//nolint:unused
	sqlCreateFuncReplIdentExists = `SELECT EXISTS (
		SELECT 1 FROM pg_proc 
		WHERE proname = 'set_replica_identity_full'
	);`
	// (replace might be brute; this checks if the repl trigger created in sqlFuncReplIfNeeded exists)

	sqlCreateEvtTriggerReplIdentExists = `SELECT EXISTS (
		SELECT 1 FROM pg_event_trigger 
		WHERE evtname = 'trg_set_replica_identity_full'
	);`

	sqlCreateEvtTriggerReplIdent = `CREATE EVENT TRIGGER trg_set_replica_identity_full ON ddl_command_end
		WHEN TAG IN ('CREATE TABLE')
		EXECUTE FUNCTION set_replica_identity_full();`
	// TIP for node reset/cleanup: DROP EVENT TRIGGER IF EXISTS trg_set_replica_identity_full;

	sqlCreatePublicationINE = `DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_publication WHERE pubname = '%[1]s'
	) THEN
		EXECUTE 'CREATE PUBLICATION %[1]s FOR ALL TABLES';
		RAISE NOTICE 'Publication %[1]s created.';
	ELSE
		RAISE NOTICE 'Publication %[1]s already exists.';
	END IF;
END$$;`

	// on startup, check for any prepared transactions and roll them back. The
	// selected columns and their order in this query is explicit so it matches
	// the preparedTxn struct.
	sqlListPreparedTxns = `SELECT transaction, gid, prepared, owner, database FROM pg_prepared_xacts;`

	sqlCreateCollationNOCASE = `CREATE COLLATION IF NOT EXISTS nocase (
		provider = icu, locale = 'und-u-ks-level2', deterministic = false
	);`

	sqlCreateUUIDExtension = `CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`

	// have to run this in a DO block because you cannot do CREATE DOMAIN IF NOT EXISTS.
	// We have to hard-code the string and then convert to numeric instead of using 2^256-1,
	// because Postgres will not precisely evaluate 2^256-1.
	sqlCreateUint256Domain = `
	DO $$ BEGIN
		CREATE DOMAIN uint256 AS NUMERIC(78)
		CHECK (VALUE >= 0 AND VALUE <= '115792089237316195423570985008687907853269984665640564039457584007913129639935'::NUMERIC(78))
		CHECK (SCALE(VALUE) = 0);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`

	// postgres returns EXTRACT as a double precision, but will only at most have 6
	// decimal places of precision (to measure microseconds). We cast to numeric(16, 6)
	// which should allow for up to 6 decimal places of precision.
	// Since max unix timestamp is 2147483648, we can cast to numeric(16, 6) to allow 10 digits
	// before the decimal and 6 after.
	sqlCreateParseUnixTimestampFunc = `CREATE OR REPLACE FUNCTION parse_unix_timestamp(timestamp_string text, format_string text)
	RETURNS NUMERIC(16, 6) AS $$
	BEGIN
		RETURN EXTRACT(EPOCH FROM TO_TIMESTAMP(timestamp_string, format_string))::numeric(16, 6);
	END;
	$$ LANGUAGE plpgsql;`

	// this is the inverse of parse_unix_timestamp
	sqlCreateFormatUnixTimestampFunc = `CREATE OR REPLACE FUNCTION format_unix_timestamp(unix_timestamp NUMERIC(16, 6), format_string text)
	RETURNS TEXT AS $$
	BEGIN
		RETURN TO_CHAR(TO_TIMESTAMP(unix_timestamp), format_string);
	END;
	$$ LANGUAGE plpgsql;`
)

func checkSuperuser(ctx context.Context, conn *pgx.Conn) error {
	user := conn.Config().User
	// Verify that the db user/role is superuser with replication privileges.
	var isSuper, isReplicator bool
	err := conn.QueryRow(ctx, `SELECT rolsuper, rolreplication FROM pg_roles WHERE rolname = $1;`, user).
		Scan(&isSuper, &isReplicator)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("postgres role does not exists: %v", user)
		}
		return fmt.Errorf("unable to verify superuser status of postgres role %v: %w", user, err)
	}
	if !isSuper || !isReplicator {
		return fmt.Errorf("postgres role is not a superuser with replication: %v", user)
	}
	return nil
}

func ensureCollation(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, sqlCreateCollationNOCASE)
	return err
}

func ensurePublication(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, fmt.Sprintf(sqlCreatePublicationINE, publicationName))
	return err
}

func ensureUUIDExtension(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, sqlCreateUUIDExtension)
	return err
}

func ensurePgCryptoExtension(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	return err
}

func ensureUint256Domain(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, sqlCreateUint256Domain)
	return err
}

func ensureUnixTimestampFuncs(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, sqlCreateParseUnixTimestampFunc)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, sqlCreateFormatUnixTimestampFunc)
	return err
}

type preparedTxn struct {
	XID      uint32    `db:"transaction"` // type xid is a 32-bit integer
	GID      string    `db:"gid"`
	Time     time.Time `db:"prepared"`
	Owner    string    `db:"owner"`
	Database string    `db:"database"`
}

func rollbackPreparedTxns(ctx context.Context, conn *pgx.Conn) (int, error) {
	rows, _ := conn.Query(ctx, sqlListPreparedTxns) // pgx ensures rows is readable and rows.Err contains any error
	preparedTxns, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[preparedTxn])
	if err != nil {
		return 0, err
	}
	var closed int
	connectedDB := conn.Config().Database
	if len(preparedTxns) > 0 {
		logger.Warnf("Found %d orphaned prepared transactions", len(preparedTxns))
	}
	for _, ptx := range preparedTxns {
		if connectedDB != ptx.Database {
			logger.Infof(`Not rolling back prepared transaction %v on foreign database %v. `+
				`A manual rollback may be required to avoid the DB hanging.`,
				ptx.GID, ptx.Database)
			continue
		}
		logger.Infof("Rolling back prepared transaction %v (xid %d) created by %v at %v",
			ptx.GID, ptx.XID, ptx.Owner, ptx.Time)
		sqlRollback := fmt.Sprintf(`ROLLBACK PREPARED '%s'`, ptx.GID)
		if _, err := conn.Exec(ctx, sqlRollback); err != nil {
			return 0, fmt.Errorf("ROLLBACK PREPARED failed: %v", err)
		}
		closed++
	}
	return closed, nil
}

const (
	InternalSchemaName = "kwild_internal"

	sentryTableName      = `sentry`
	sentryTableNameFull  = InternalSchemaName + "." + sentryTableName
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
	// First create the function if needed.
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
	exists, err := tableExists(ctx, InternalSchemaName, sentryTableName, conn)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	createStmt := fmt.Sprintf(sqlCreateSchemaTemplate, InternalSchemaName)
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
