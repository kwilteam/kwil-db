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

	sqlCreateUint256Domain = `CREATE DOMAIN uint256 AS NUMERIC(78) NOT NULL
	CHECK (VALUE >= 0 AND VALUE < 2^256)
	CHECK (SCALE(VALUE) = 0);
	;`
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

func ensureUint256Domain(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, sqlCreateUint256Domain)
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
	logger.Warnf("Found %d orphaned prepared transactions", len(preparedTxns))
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
	return closed, err
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
