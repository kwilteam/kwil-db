package execution

import (
	"context"
	"encoding/json"
	"fmt"

	_ "embed"

	"github.com/kwilteam/kwil-db/common"
	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/ddl"
	procedural "github.com/kwilteam/kwil-db/internal/engine/procedures"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
)

var (
	// engineVersion is the version of the 'kwild_internal' schema
	engineVersion int64 = 0

	schemaVersion        = 0 // schema version allows upgrading schemas in the future
	sqlCreateSchemaTable = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.kwil_schemas (
	dbid TEXT PRIMARY KEY,
	schema_content BYTEA,
	version INT DEFAULT %d
);`, pg.InternalSchemaName, schemaVersion)
	sqlCreateSchema    = `CREATE SCHEMA "%s";`
	sqlStoreKwilSchema = fmt.Sprintf(`INSERT INTO %s.kwil_schemas (dbid, schema_content, version) VALUES ($1, $2, $3)
	ON CONFLICT (dbid) DO UPDATE SET schema_content = $2, version = $3`, pg.InternalSchemaName)
	sqlListSchemaContent = fmt.Sprintf(`SELECT schema_content FROM %s.kwil_schemas;`, pg.InternalSchemaName)
	sqlDropSchema        = `DROP SCHEMA "%s" CASCADE;`
	sqlDeleteKwilSchema  = fmt.Sprintf(`DELETE FROM %s.kwil_schemas WHERE dbid = $1;`, pg.InternalSchemaName)
)

func initTables(ctx context.Context, db sql.DB) error {
	if err := createSchemasTableIfNotExists(ctx, db); err != nil {
		return err
	}
	return nil
}

func dbidSchema(dbid string) string {
	return pg.DefaultSchemaFilterPrefix + dbid
}

// createSchemasTableIfNotExists creates the schemas table if it does not exist
func createSchemasTableIfNotExists(ctx context.Context, tx sql.DB) error {
	_, err := tx.Execute(ctx, sqlCreateSchemaTable)
	return err
}

// createSchema creates a schema in the database.
// It will also store the schema in the kwil_schemas table.
// It also creates the relevant tables, indexes, etc.
// If the schema already exists in the Kwil schemas table, it will be updated.
func createSchema(ctx context.Context, tx sql.TxMaker, schema *types.Schema) error {
	schemaName := dbidSchema(schema.DBID())

	sp, err := tx.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer sp.Rollback(ctx)

	_, err = sp.Execute(ctx, fmt.Sprintf(sqlCreateSchema, schemaName))
	if err != nil {
		return err
	}

	// we can json marshal without concern for non-determinism
	// because kwil_schemas exists outside of consensus / replicated state
	schemaBts, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	// since we will fail if the schema already exists, we can assume that it does not exist
	// in the kwil_schemas table. If it does for some reason, we will update it.
	_, err = sp.Execute(ctx, sqlStoreKwilSchema, schema.DBID(), schemaBts, schemaVersion)
	if err != nil {
		return err
	}

	for _, table := range schema.Tables {
		statements, err := ddl.GenerateDDL(schemaName, table)
		if err != nil {
			return err
		}

		for _, stmt := range statements {
			_, err = sp.Execute(ctx, stmt)
			if err != nil {
				return err
			}
		}
	}

	// for each procedure, we will sanitize it,
	// type check, generate the PLPGSQL code,
	// and then execute the generated code.
	stmts, err := procedural.GeneratePLPGSQL(schema, schemaName, pgSessionPrefix, PgSessionVars, &procedural.GenerateOptions{
		LogProcedureNameOnError: true,
	})
	if err != nil {
		return err
	}
	for _, stmt := range stmts {
		_, err = sp.Execute(ctx, stmt)
		if err != nil {
			return err
		}
	}

	return sp.Commit(ctx)
}

// getSchemas returns all schemas in the kwil_schemas table
func getSchemas(ctx context.Context, tx sql.Executor) ([]*types.Schema, error) {
	res, err := tx.Execute(ctx, sqlListSchemaContent)
	if err != nil {
		return nil, err
	}

	schemas := make([]*types.Schema, len(res.Rows))
	for i, row := range res.Rows {
		if len(row) != 1 {
			return nil, fmt.Errorf("expected 1 column, got %d", len(row))
		}

		bts, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("expected []byte, got %T", row[0])
		}

		schema := &types.Schema{}
		err := json.Unmarshal(bts, schema)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling schema: %w", err)
		}

		schemas[i] = schema
	}

	return schemas, nil
}

// deleteSchema deletes a schema from the database.
// It will also delete the schema from the kwil_schemas table.
func deleteSchema(ctx context.Context, tx sql.TxMaker, dbid string) error {
	schemaName := dbidSchema(dbid)

	sp, err := tx.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer sp.Rollback(ctx)

	_, err = sp.Execute(ctx, fmt.Sprintf(sqlDropSchema, schemaName))
	if err != nil {
		return err
	}

	_, err = sp.Execute(ctx, sqlDeleteKwilSchema, dbid)
	if err != nil {
		return err
	}

	return sp.Commit(ctx)
}

// setContextualVars sets the contextual variables for the given postgres session.
func setContextualVars(ctx context.Context, db sql.DB, data *common.ExecutionData) error {
	// for contextual parameters, we use postgres's current_setting()
	// feature for setting session variables. For example, @caller
	// is accessed via current_setting('ctx.caller')

	_, err := db.Execute(ctx, fmt.Sprintf(`SET LOCAL %s.%s = '%s';`, pgSessionPrefix, callerVar, data.Caller))
	if err != nil {
		return err
	}

	_, err = db.Execute(ctx, fmt.Sprintf(`SET LOCAL %s.%s = '%s';`, pgSessionPrefix, txidVar, data.TxID))
	if err != nil {
		return err
	}

	return nil
}

var (
	pgSessionPrefix = "ctx"
	callerVar       = "caller"
	txidVar         = "txid"
	PgSessionVars   = map[string]*types.DataType{
		callerVar: types.TextType,
		txidVar:   types.TextType,
	}
)
