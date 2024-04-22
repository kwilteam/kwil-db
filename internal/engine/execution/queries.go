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
	engineVersion int64 = 1

	schemaVersion        = 0 // schema version allows upgrading schemas in the future
	sqlCreateSchemaTable = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.kwil_schemas (
	dbid TEXT PRIMARY KEY,
	schema_content BYTEA,
	version INT DEFAULT %d
);`, pg.InternalSchemaName, schemaVersion)
	sqlCreateSchema    = `CREATE SCHEMA "%s";`
	sqlStoreKwilSchema = fmt.Sprintf(`INSERT INTO %s.kwil_schemas (id, dbid, schema_content, version, owner, name)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (dbid) DO UPDATE SET schema_content = $3, version = $4, owner = $5, name = $6;`, pg.InternalSchemaName)
	sqlStoreProcedure = fmt.Sprintf(`INSERT INTO %s.procedures (name, schema_id, param_types, param_names, return_types, return_names, returns_table, public, owner_only, is_view)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`, pg.InternalSchemaName)
	sqlListSchemaContent = fmt.Sprintf(`SELECT schema_content FROM %s.kwil_schemas;`, pg.InternalSchemaName)
	sqlDropSchema        = `DROP SCHEMA "%s" CASCADE;`
	sqlDeleteKwilSchema  = fmt.Sprintf(`DELETE FROM %s.kwil_schemas WHERE dbid = $1;`, pg.InternalSchemaName)

	// v1 upgrades the schema to be:
	// TABLE kwil_schemas (
	// ID uuid PRIMARY KEY,
	// 	dbid TEXT,
	// 	schema_content BYTEA,
	// 	version INT DEFAULT 0,
	// 	owner BYTEA,
	// 	name TEXT
	// )
	// TABLE procedures (
	// 	name TEXT,
	// 	schema_id UUID,
	//  param_types TEXT[],
	//  param_names TEXT[],
	//  return_types TEXT[],
	//  return_names TEXT[],
	//  returns_table BOOLEAN,
	//  public BOOLEAN,
	//  owner_only BOOLEAN,
	//  is_view BOOLEAN,
	//  primary key (name, schema)
	//  FOREIGN KEY (schema) REFERENCES kwil_schemas (id) ON UPDATE CASCADE ON DELETE CASCADE
	//

	// upgrades for v1:
	// to change the primary key, we will:
	// 1. add a new uuid column, and generate a uuid for each schema
	// 2. remove the old primary key
	// 3. add a new primary key
	sqlUpgradeSchemaTableV1AddUUIDColumn = fmt.Sprintf(`
	ALTER TABLE %s.kwil_schemas ADD COLUMN id UUID;
	`, pg.InternalSchemaName)
	// sqlBackfillSchemaTableV1UUID adds a UUID to all existing schemas.
	// It uses a random UUID namespace to generate the UUIDs from the dbid.
	// This namespace is not used anywhere else. We want to decouple the UUID
	// from the DBID, so separate UUIDs will be used in the future that are based on
	// the txid.
	sqlBackfillSchemaTableV1UUID = fmt.Sprintf(`
	UPDATE %s.kwil_schemas SET id = uuid_generate_v5('052d10c4-acf8-4ec9-a616-105bf1d1e873'::uuid, dbid);
	`, pg.InternalSchemaName)
	sqlUpgradeRemovePrimaryKey = fmt.Sprintf(`
	ALTER TABLE %s.kwil_schemas DROP CONSTRAINT kwil_schemas_pkey;
	`, pg.InternalSchemaName)
	sqlUpgradeAddPrimaryKeyV1UUID = fmt.Sprintf(`
	ALTER TABLE %s.kwil_schemas ADD PRIMARY KEY (id);
	`, pg.InternalSchemaName)

	// unique constraint for the dbid
	sqlUpgradeAddUniqueConstraintV1DBID = fmt.Sprintf(`
	ALTER TABLE %s.kwil_schemas ADD CONSTRAINT kwil_schemas_dbid_unique UNIQUE (dbid);
	`, pg.InternalSchemaName)

	sqlUpgradeSchemaTableV1AddOwnerColumn = fmt.Sprintf(`
	ALTER TABLE %s.kwil_schemas ADD COLUMN name TEXT;
	`, pg.InternalSchemaName)
	sqlUpgradeSchemaTableV1AddNameColumn = fmt.Sprintf(`
	ALTER TABLE %s.kwil_schemas ADD COLUMN owner BYTEA;
	`, pg.InternalSchemaName)
	// sqlBackfillSchemaTableV1 adds the owner and name to all existing schemas,
	// and updates the version to 1.
	sqlBackfillSchemaTableV1 = fmt.Sprintf(`
	UPDATE %s.kwil_schemas SET owner = $1, name = $2, version = 1;
	`, pg.InternalSchemaName)

	sqlAddProceduresTableV1 = fmt.Sprintf(`
	CREATE TABLE %s.procedures (
		name TEXT not null,
		schema_id uuid not null,
		param_types TEXT[],
		param_names TEXT[],
		return_types TEXT[],
		return_names TEXT[],
		returns_table BOOLEAN not null,
		public BOOLEAN not null,
		owner_only BOOLEAN not null,
		is_view BOOLEAN not null,
		primary key (name, schema_id),
		FOREIGN KEY (schema_id) REFERENCES %s.kwil_schemas (id) ON UPDATE CASCADE ON DELETE CASCADE
	)
	`, pg.InternalSchemaName, pg.InternalSchemaName)
	sqlIndexProceduresTableV1SchemaID = fmt.Sprintf(`
	CREATE INDEX procedures_schema_id ON %s.procedures (schema_id);
	`, pg.InternalSchemaName)
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

// a random uuidNamespace for generating UUIDs for schemas.
var uuidNamespace = "01c32544-c21f-4522-98c5-40e6fb0a0831"

// createSchema creates a schema in the database.
// It will also store the schema in the kwil_schemas table.
// It also creates the relevant tables, indexes, etc.
// If the schema already exists in the Kwil schemas table, it will be updated.
func createSchema(ctx context.Context, tx sql.TxMaker, schema *types.Schema, txid string) error {
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

	uuidNamespace, err := types.ParseUUID(uuidNamespace)
	if err != nil {
		return err
	}

	uuid := types.NewUUIDV5WithNamespace(*uuidNamespace, []byte(txid))

	// since we will fail if the schema already exists, we can assume that it does not exist
	// in the kwil_schemas table. If it does for some reason, we will update it.
	_, err = sp.Execute(ctx, sqlStoreKwilSchema, uuid, schema.DBID(), schemaBts, schemaVersion, schema.Owner, schema.Name)
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

	// store the procedures in the kwil_procedures table
	for _, proc := range schema.Procedures {

		var paramTypes []string
		var paramNames []string
		for _, col := range proc.Parameters {
			paramTypes = append(paramTypes, col.Type.String())
			paramNames = append(paramNames, col.Name)
		}

		var returnTypes []string
		var returnNames []string
		returnsTable := false
		if proc.Returns != nil {
			returnsTable = proc.Returns.IsTable
			for _, col := range proc.Returns.Fields {
				returnTypes = append(returnTypes, col.Type.String())
				returnNames = append(returnNames, col.Name)
			}
		}

		_, err = sp.Execute(ctx, sqlStoreProcedure,
			proc.Name,
			uuid,
			paramTypes,
			paramNames,
			returnTypes,
			returnNames,
			returnsTable,
			proc.Public,
			proc.IsOwner(),
			proc.IsView())
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
