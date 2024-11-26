package interpreter

import (
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/parse"
)

const pgSchema = `kwild_public` // TODO: remove me, since we are now multi-schema

var (
	//go:embed schema.sql
	schemaInitSQL string
)

// initSQL initializes the SQL schema.
func initSQL(ctx context.Context, db sql.DB) error {
	return pg.Exec(ctx, db, schemaInitSQL)
}

// returnsExactlyOneInt64 ensures a result set has exactly one row and column.
func returnsExactlyOneInt64(rows *sql.ResultSet) (int64, error) {
	if len(rows.Rows) != 1 {
		return 0, errors.New("expected exactly one row")
	}
	if len(rows.Rows[0]) != 1 {
		return 0, errors.New("expected exactly one column")
	}

	t, ok := sql.Int64(rows.Rows[0][0])
	if ok {
		return t, nil
	}

	return 0, fmt.Errorf("expected int64, got %T", rows.Rows[0][0])
}

// createUserNamespace creates a new schema for a user.
func createUserNamespace(ctx context.Context, db sql.DB, name string, owner []byte) error {
	_, err := db.Execute(ctx, `CREATE SCHEMA `+name)
	if err != nil {
		return err
	}

	_, err = db.Execute(ctx, `INSERT INTO kwild_engine.user_namespaces (name, owner) VALUES ($1, $2)`, name, owner)
	return err
}

// storeAction stores an action in the database.
// It should always be called within a transaction.
func storeAction(ctx context.Context, db sql.DB, namespace string, action *Action) error {
	returnsTable := false
	if action.Returns != nil {
		returnsTable = action.Returns.IsTable
	}

	modStrs := make([]string, len(action.Modifiers))
	for i, mod := range action.Modifiers {
		modStrs[i] = string(mod)
	}

	res, err := db.Execute(ctx, `INSERT INTO kwild_engine.actions (name, schema_name, public, raw_body, modifiers, returns_table)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		action.Name, namespace, action.Public, action.RawBody, modStrs, returnsTable)
	if err != nil {
		return err
	}
	actionID, err := returnsExactlyOneInt64(res)
	if err != nil {
		return err
	}

	for _, param := range action.Parameters {
		dt, err := param.Type.PGScalar()
		if err != nil {
			return err
		}

		_, err = db.Execute(ctx, `INSERT INTO kwild_engine.parameters (action_id, name, scalar_type, is_array, metadata)
			VALUES ($1, $2, $3, $4, $5)`,
			actionID, param.Name, dt, param.Type.IsArray, getTypeMetadata(param.Type))
		if err != nil {
			return err
		}
	}

	if action.Returns != nil {
		for _, field := range action.Returns.Fields {
			dt, err := field.Type.PGScalar()
			if err != nil {
				return err
			}

			_, err = db.Execute(ctx, `INSERT INTO kwild_engine.return_fields (action_id, name, scalar_type, is_array, metadata)
			VALUES ($1, $2, $3, $4, $5)`,
				actionID, field.Name, dt, field.Type.IsArray, getTypeMetadata(field.Type))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// loadActions loads all actions from the database.
// it maps: schema -> action name -> action
func loadActions(ctx context.Context, db sql.DB) (map[string]map[string]*Action, error) {
	schemas := make(map[string]map[string]*Action)

	var id int64
	var schemaName, name, rawBody string
	var public, returnsTable bool
	var modifiers []string
	scans := []any{
		&id,
		&schemaName,
		&name,
		&public,
		&rawBody,
		&modifiers,
		&returnsTable,
	}

	err := pg.QueryRowFunc(ctx, db, `SELECT id, schema_name, name, public, raw_body, modifiers::TEXT[], returns_table FROM kwild_engine.actions`,
		scans, func() error {
			action := &Action{
				Name:    name,
				Public:  public,
				RawBody: rawBody,
			}

			schema, ok := schemas[schemaName]
			if !ok {
				schema = make(map[string]*Action)
				schemas[schemaName] = schema
			}

			for _, mod := range modifiers {
				action.Modifiers = append(action.Modifiers, Modifier(mod))
			}

			res, err := parse.ParseActionBodyWithoutValidation(action.RawBody)
			if err != nil {
				return err
			}
			if res.Errs.Err() != nil {
				return res.Errs.Err()
			}
			action.Body = res.AST

			if returnsTable {
				action.Returns = &ActionReturn{
					IsTable: true,
					Fields:  nil,
				}
			}

			schema[action.Name] = action
			return nil
		})
	if err != nil {
		return nil, err
	}

	var actionName, paramName, paramScalarType string
	var paramIsArray bool
	var paramMetadata []byte
	scans = []any{
		&actionName,
		&schemaName,
		&paramName,
		&paramScalarType,
		&paramIsArray,
		&paramMetadata,
	}

	err = pg.QueryRowFunc(ctx, db, `SELECT a.name, a.schema_name, p.name, p.scalar_type::TEXT, p.is_array, p.metadata FROM kwild_engine.parameters p
		JOIN kwild_engine.actions a ON a.id = action_id`, scans, func() error {
		action, ok := schemas[schemaName][actionName]
		if !ok {
			// suggests an internal error
			return fmt.Errorf("action %s not found", actionName)
		}

		dt, err := types.ParseDataType(paramScalarType)
		if err != nil {
			return err
		}

		if paramMetadata != nil {
			dt.Metadata = reconstructTypeMetadata(paramMetadata)
		}

		dt.IsArray = paramIsArray

		action.Parameters = append(action.Parameters, &NamedType{
			Name: paramName,
			Type: dt,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	// technically redundant with the scans slice defined above, but it's clearer this way,
	// just to show what this query expects
	scans = []any{
		&actionName,
		&schemaName,
		&paramName,
		&paramScalarType,
		&paramIsArray,
		&paramMetadata,
	}

	err = pg.QueryRowFunc(ctx, db, `SELECT a.name, a.schema_name, p.name, p.scalar_type::TEXT, p.is_array, p.metadata FROM kwild_engine.return_fields p
		JOIN kwild_engine.actions a ON a.id = action_id`, scans, func() error {
		action, ok := schemas[schemaName][actionName]
		if !ok {
			// suggests an internal error
			return fmt.Errorf("action %s not found", actionName)
		}

		// if the action doesn't have a return type, create one.
		// Up until this point, we only knew if an action had a return if it returned a table.
		// But if a row is returned here for an action, that means it has a return type.
		if action.Returns == nil {
			action.Returns = &ActionReturn{
				IsTable: false,
			}
		}

		dt, err := types.ParseDataType(paramScalarType)
		if err != nil {
			return err
		}

		if paramMetadata != nil {
			dt.Metadata = reconstructTypeMetadata(paramMetadata)
		}

		dt.IsArray = paramIsArray

		action.Returns.Fields = append(action.Returns.Fields, &NamedType{
			Name: paramName,
			Type: dt,
		})

		return nil
	})

	return schemas, err
}

// listNamespaceTables lists all namespaces that are created by users.
// It returns a mapping of schemas -> table names -> tables.
func listNamespaceTables(ctx context.Context, db sql.DB) (map[string]map[string]*engine.Table, error) {
	schemas := make(map[string]map[string]*engine.Table)
	var schemaName string
	var tblName string
	var colNames, dataTypes, indexNames, constraintNames, constraintTypes, fkNames, fkOnUpdate, fkOnDelete []string
	var indexCols, constraintCols, fkCols [][]string
	var isNullables, isPrimaryKeys, isPKs, isUniques []bool
	scans := []any{
		&schemaName,
		&tblName,
		&colNames,
		&dataTypes,
		&isNullables,
		&isPrimaryKeys,
		&indexNames,
		&isPKs,
		&isUniques,
		&indexCols,
		&constraintNames,
		&constraintTypes,
		&constraintCols,
		&fkNames,
		&fkCols,
		&fkOnUpdate,
		&fkOnDelete,
	}
	err := pg.QueryRowFunc(ctx, db, `
	WITH columns AS (
		SELECT c.schema_name, c.table_name,
			array_agg(c.column_name ORDER BY c.ordinal_position) AS column_names,
			array_agg(c.data_type ORDER BY c.ordinal_position) AS data_types,
			array_agg(c.is_nullable ORDER BY c.ordinal_position) AS is_nullables,
			array_agg(c.is_primary_key ORDER BY c.ordinal_position) AS is_primary_keys
		FROM kwil_columns c
		GROUP BY c.schema_name, c.table_name
	),
	indexes AS (
		SELECT i.schema_name, i.table_name,
			array_agg(i.index_name ORDER BY i.index_name) AS index_names,
			array_agg(i.is_pk ORDER BY i.index_name) AS is_pks,
			array_agg(i.is_unique ORDER BY i.index_name) AS is_uniques,
			array_agg(i.column_names ORDER BY i.index_name) AS column_names
		FROM kwil_indexes i
		GROUP BY i.schema_name, i.table_name
	), constraints AS (
		SELECT c.schema_name, c.table_name,
			array_agg(c.constraint_name ORDER BY c.constraint_name) AS constraint_names,
			array_agg(c.constraint_type ORDER BY c.constraint_name) AS constraint_types,
			array_agg(c.columns ORDER BY c.constraint_name) AS columns
		FROM kwil_constraints c
		GROUP BY c.schema_name, c.table_name
	), foreign_keys AS (
		SELECT f.schema_name, f.table_name,
			array_agg(f.constraint_name ORDER BY f.constraint_name) AS constraint_names,
			array_agg(f.columns ORDER BY f.constraint_name) AS columns,
			array_agg(f.on_update ORDER BY f.constraint_name) AS on_updates,
			array_agg(f.on_delete ORDER BY f.constraint_name) AS on_deletes
		FROM kwil_foreign_keys f
		GROUP BY f.schema_name, f.table_name
	)
	SELECT 
		t.schema, t.name,
		c.column_names, c.data_types, c.is_nullables, c.is_primary_keys,
		i.index_names, i.is_pks, i.is_uniques, i.column_names,
		co.constraint_names, co.constraint_types, co.columns,
		f.constraint_names, f.columns, f.on_updates, f.on_deletes
	FROM kwil_tables t
	JOIN columns c ON t.name = c.table_name AND t.schema = c.schema_name
	LEFT JOIN indexes i ON t.name = i.table_name AND t.schema = i.schema_name
	LEFT JOIN constraints co ON t.name = co.table_name AND t.schema = co.schema_name
	LEFT JOIN foreign_keys f ON t.name = f.table_name AND t.schema = f.schema_name
		`, scans,
		func() error {
			schema, ok := schemas[schemaName]
			if !ok {
				schema = make(map[string]*engine.Table)
				schemas[schemaName] = schema
			}

			_, ok = schema[tblName]
			if ok {
				// some basic validation
				return fmt.Errorf("duplicate table %s", tblName)
			}

			tbl := &engine.Table{
				Name:        tblName,
				Constraints: make(map[string]*engine.Constraint),
			}
			schema[tblName] = tbl

			// add columns
			for i, colName := range colNames {
				dt, err := types.ParseDataType(dataTypes[i])
				if err != nil {
					return err
				}

				tbl.Columns = append(tbl.Columns, &engine.Column{
					Name:         colName,
					DataType:     dt,
					Nullable:     isNullables[i],
					IsPrimaryKey: isPrimaryKeys[i],
				})
			}

			// add indexes
			for i, indexName := range indexNames {
				indexType := engine.BTREE
				if isPKs[i] {
					indexType = engine.PRIMARY
				} else if isUniques[i] {
					indexType = engine.UNIQUE_BTREE
				}

				tbl.Indexes = append(tbl.Indexes, &engine.Index{
					Name:    indexName,
					Columns: indexCols[i],
					Type:    indexType,
				})
			}

			// add constraints
			for i, constraintName := range constraintNames {
				var constraintType engine.ConstraintType
				switch strings.ToLower(constraintTypes[i]) {
				case "unique":
					constraintType = engine.ConstraintUnique
				case "check":
					constraintType = engine.ConstraintCheck
				default:
					return fmt.Errorf("unknown constraint type %s", constraintTypes[i])
				}

				_, ok := tbl.Constraints[constraintName]
				if ok {
					return fmt.Errorf("duplicate constraint %s", constraintName)
				}

				tbl.Constraints[constraintName] = &engine.Constraint{
					Name:    constraintName,
					Type:    constraintType,
					Columns: constraintCols[i],
				}
			}

			// add foreign keys
			for i, fkName := range fkNames {
				_, ok := tbl.Constraints[fkName]
				if ok {
					return fmt.Errorf("duplicate foreign key %s", fkName)
				}

				fk := &engine.Constraint{
					Name:    fkName,
					Type:    engine.ConstraintFK,
					Columns: fkCols[i],
				}

				tbl.Constraints[fkName] = fk
			}

			schemas[schemaName][tblName] = tbl
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return schemas, nil
}

// getNamespaceOwners returns the owners of the given namespaces.
// It maps the namespace to the owner's public key.
func getNamespaceOwners(ctx context.Context, db sql.DB) (map[string][]byte, error) {
	owners := make(map[string][]byte)
	var name string
	var owner []byte
	scans := []any{
		&name,
		&owner,
	}

	err := pg.QueryRowFunc(ctx, db, `SELECT name, owner FROM kwild_engine.user_namespaces`, scans, func() error {
		owners[name] = owner
		return nil
	})
	if err != nil {
		return nil, err
	}

	return owners, nil
}

// namedTypeFromRow creates a NamedType from a row.
func namedTypeFromRow(row []interface{}) (*NamedType, error) {
	if len(row) != 4 {
		return nil, fmt.Errorf("expected 4 columns, got %d", len(row))
	}

	dt := &types.DataType{
		Name:    row[1].(string),
		IsArray: row[2].(bool),
	}

	switch meta := row[3].(type) {
	case nil:
		// no metadata
	case []byte:
		dt.Metadata = reconstructTypeMetadata(meta)
	default:
		return nil, fmt.Errorf("unexpected metadata type %T", meta)
	}

	return &NamedType{
		Name: row[0].(string),
		Type: dt,
	}, nil
}

// getTypeMetadata gets the serialized type metadata.
// If there is none, it returns nil.
func getTypeMetadata(t *types.DataType) []byte {
	if t.Metadata == nil {
		return nil
	}

	meta := make([]byte, 4)
	binary.LittleEndian.PutUint16(meta[:2], t.Metadata[0])
	binary.LittleEndian.PutUint16(meta[2:], t.Metadata[1])

	return meta
}

// reconstructTypeMetadata reconstructs the type metadata from the serialized form.
func reconstructTypeMetadata(meta []byte) *[2]uint16 {
	if len(meta) == 0 {
		return nil
	}

	return &[2]uint16{
		binary.LittleEndian.Uint16(meta[:2]),
		binary.LittleEndian.Uint16(meta[2:]),
	}
}

// query executes a SQL query with the given values.
// It is a utility function to help reduce boilerplate when executing
// SQL with Value types.
func query(ctx context.Context, db sql.DB, query string, scanVals []Value, fn func() error, args []Value) error {
	argVals := make([]any, len(args))
	var err error
	for i, v := range args {
		argVals[i], err = v.DBValue()
		if err != nil {
			return err
		}
	}

	recVals := make([]any, len(scanVals))
	for i := range scanVals {
		recVals[i], err = scanVals[i].DBValue()
		if err != nil {
			return err
		}
	}

	return pg.QueryRowFunc(ctx, db, query, recVals, fn, argVals)
}
