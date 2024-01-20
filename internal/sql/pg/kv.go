package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// This file defines queries for a kv store emulated as a BYTEA-BYTEA table in
// SQL, used by the DB type's Set/Get methods.

// [dev note]: either the table name must be prefixed with a pg db schema, or the
// key is a concatenation of namespace (e.g. dbid) and another semantic key. For
// now, the statements below are defined for the latter approach, which means
// that there is a single kv table in the "kwild_internal" postgresql schema.

const (
	kvTableName     = "kv"
	kvTableNameFull = internalSchemaName + "." + kvTableName

	createKvStmt = `
		CREATE TABLE IF NOT EXISTS ` + kvTableNameFull + ` (
			key BYTEA PRIMARY KEY,
			value BYTEA NOT NULL
		);
	`
	createKvStmtTmpl = `CREATE TABLE IF NOT EXISTS %s (key BYTEA PRIMARY KEY, value BYTEA NOT NULL);`

	insertKvStmt = `
		INSERT INTO ` + kvTableNameFull + ` (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2;
	`
	insertKvStmtTmpl = `INSERT INTO %s (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2;`

	selectKvStmt = `
		SELECT value
		FROM ` + kvTableNameFull + `
		WHERE key = $1;
	`
	selectKvStmtTmpl = `SELECT value FROM %s WHERE key = $1;`

	deleteKvStmt = `
		DELETE FROM ` + kvTableNameFull + `
		WHERE key = $1;
	`
	deleteKvStmtTmpl = `DELETE FROM %s WHERE key = $1;`
)

func ensureKvTable(ctx context.Context, conn *pgx.Conn) error {
	exists, err := tableExists(ctx, internalSchemaName, kvTableName, conn)
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
	_, err = conn.Exec(ctx, createKvStmt)
	return err
}
