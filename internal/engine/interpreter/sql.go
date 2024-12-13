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

// createNamespace creates a new schema for a user.
func createNamespace(ctx context.Context, db sql.DB, name string, nsType namespaceType) error {
	_, err := db.Execute(ctx, `CREATE SCHEMA `+name)
	if err != nil {
		return err
	}

	_, err = db.Execute(ctx, `INSERT INTO kwild_engine.namespaces (name, type) VALUES ($1, $2)`, name, nsType)
	return err
}

// dropNamespace drops a schema for a user.
func dropNamespace(ctx context.Context, db sql.DB, name string) error {
	_, err := db.Execute(ctx, `DROP SCHEMA `+name+` CASCADE`)
	if err != nil {
		return err
	}

	_, err = db.Execute(ctx, `DELETE FROM kwild_engine.namespaces WHERE name = $1`, name)
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

	res, err := db.Execute(ctx, `INSERT INTO kwild_engine.actions (name, schema_name, public, raw_statement, modifiers, returns_table)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		action.Name, namespace, action.Public, action.RawStatement, modStrs, returnsTable)
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

// listNamespaces lists all namespaces that are created.
func listNamespaces(ctx context.Context, db sql.DB) ([]struct {
	Name string
	Type namespaceType
}, error) {
	var namespaces []struct {
		Name string
		Type namespaceType
	}
	var namespace string
	var nsType string
	err := pg.QueryRowFunc(ctx, db, `SELECT name, type::TEXT FROM kwild_engine.namespaces`, []any{&namespace, &nsType},
		func() error {
			nsT := namespaceType(nsType)
			if !nsT.valid() {
				return fmt.Errorf("unknown namespace type %s", nsType)
			}

			namespaces = append(namespaces, struct {
				Name string
				Type namespaceType
			}{Name: namespace, Type: nsT})
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return namespaces, nil
}

// listTablesInNamespace lists all tables in a namespace.
func listTablesInNamespace(ctx context.Context, db sql.DB, namespace string) ([]*engine.Table, error) {
	tables := make([]*engine.Table, 0)
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
		FROM info.columns c
		GROUP BY c.schema_name, c.table_name
	),
	indexes AS (
		SELECT i.schema_name, i.table_name,
			array_agg(i.index_name ORDER BY i.index_name) AS index_names,
			array_agg(i.is_pk ORDER BY i.index_name) AS is_pks,
			array_agg(i.is_unique ORDER BY i.index_name) AS is_uniques,
			array_agg(i.column_names ORDER BY i.index_name) AS column_names
		FROM info.indexes i
		GROUP BY i.schema_name, i.table_name
	), constraints AS (
		SELECT c.schema_name, c.table_name,
			array_agg(c.constraint_name ORDER BY c.constraint_name) AS constraint_names,
			array_agg(c.constraint_type ORDER BY c.constraint_name) AS constraint_types,
			array_agg(c.columns ORDER BY c.constraint_name) AS columns
		FROM info.constraints c
		GROUP BY c.schema_name, c.table_name
	), foreign_keys AS (
		SELECT f.schema_name, f.table_name,
			array_agg(f.constraint_name ORDER BY f.constraint_name) AS constraint_names,
			array_agg(f.columns ORDER BY f.constraint_name) AS columns,
			array_agg(f.on_update ORDER BY f.constraint_name) AS on_updates,
			array_agg(f.on_delete ORDER BY f.constraint_name) AS on_deletes
		FROM info.foreign_keys f
		GROUP BY f.schema_name, f.table_name
	)
	SELECT
		t.schema, t.name,
		c.column_names, c.data_types, c.is_nullables, c.is_primary_keys,
		i.index_names, i.is_pks, i.is_uniques, i.column_names,
		co.constraint_names, co.constraint_types, co.columns,
		f.constraint_names, f.columns, f.on_updates, f.on_deletes
	FROM info.tables t
	JOIN columns c ON t.name = c.table_name AND t.schema = c.schema_name
	LEFT JOIN indexes i ON t.name = i.table_name AND t.schema = i.schema_name
	LEFT JOIN constraints co ON t.name = co.table_name AND t.schema = co.schema_name
	LEFT JOIN foreign_keys f ON t.name = f.table_name AND t.schema = f.schema_name
	WHERE t.schema = $1`, scans,
		func() error {
			tbl := &engine.Table{
				Name:        tblName,
				Constraints: make(map[string]*engine.Constraint),
			}

			tables = append(tables, tbl)

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
					Type:    engine.ConstraintFK,
					Columns: fkCols[i],
				}

				tbl.Constraints[fkName] = fk
			}
			return nil
		}, namespace,
	)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

// listActionsInNamespace lists all actions in a namespace.
func listActionsInNamespace(ctx context.Context, db sql.DB, namespace string) ([]*Action, error) {
	var actions []*Action
	var rawStmt string
	scans := []any{
		&rawStmt,
	}

	err := pg.QueryRowFunc(ctx, db, `SELECT raw_statement FROM kwild_engine.actions WHERE schema_name = $1`, scans,
		func() error {
			res, err := parse.Parse(rawStmt)
			if err != nil {
				return err
			}

			if len(res) != 1 {
				return fmt.Errorf("expected exactly 1 statement, got %d", len(res))
			}

			createActionStmt, ok := res[0].(*parse.CreateActionStatement)
			if !ok {
				return fmt.Errorf("expected CreateActionStatement, got %T", res[0])
			}

			act := &Action{}
			err = act.FromAST(createActionStmt)
			if err != nil {
				return err
			}

			actions = append(actions, act)
			return nil
		}, namespace,
	)
	if err != nil {
		return nil, err
	}

	return actions, nil
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

// query executes a SQL query with the given values.
// It is a utility function to help reduce boilerplate when executing
// SQL with Value types.
func query(ctx context.Context, db sql.DB, query string, scanVals []Value, fn func() error, args []Value) error {
	argVals := make([]any, len(args))
	var err error
	for i, v := range args {
		argVals[i] = v
	}

	recVals := make([]any, len(scanVals))
	for i := range scanVals {
		recVals[i] = scanVals[i]
	}

	err = pg.QueryRowFunc(ctx, db, query, recVals, fn, argVals...)
	if err != nil {
		return err
	}

	return nil
}
