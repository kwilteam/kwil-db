package datasets

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/sql/driver"
	"strings"
)

type metadataTable string

var metadataTables = []metadataTable{
	tablesTable,
	columnsTable,
	attributesTable,
	indexesTable,
	actionsTable,
	statementsTable,
}

const (
	tablesTable     metadataTable = "_tables"
	columnsTable    metadataTable = "_columns"
	attributesTable metadataTable = "_attributes"
	indexesTable    metadataTable = "_indexes"
	actionsTable    metadataTable = "_actions"
	statementsTable metadataTable = "_action_statements"
)

func (m metadataTable) initStmt() string {
	switch m {
	case tablesTable:
		return "CREATE TABLE IF NOT EXISTS _tables (id INTEGER PRIMARY KEY AUTOINCREMENT, table_name TEXT NOT NULL);"
	case columnsTable:
		return "CREATE TABLE IF NOT EXISTS _columns (id INTEGER PRIMARY KEY AUTOINCREMENT, table_id INTEGER NOT NULL, column_name TEXT NOT NULL, column_type INTEGER NOT NULL, FOREIGN KEY (table_id) REFERENCES _tables(id));"
	case attributesTable:
		return "CREATE TABLE IF NOT EXISTS _attributes (id INTEGER PRIMARY KEY AUTOINCREMENT, column_id INTEGER NOT NULL, attribute_type INTEGER NOT NULL, attribute_value BLOB, FOREIGN KEY (column_id) REFERENCES _columns(id));"
	case indexesTable:
		return "CREATE TABLE IF NOT EXISTS _indexes (id INTEGER PRIMARY KEY AUTOINCREMENT, table_id INTEGER NOT NULL, index_name TEXT NOT NULL, index_type INTEGER NOT NULL, columns TEXT NOT NULL, FOREIGN KEY (table_id) REFERENCES _tables(id));"
	case actionsTable:
		return "CREATE TABLE IF NOT EXISTS _actions (id INTEGER PRIMARY KEY AUTOINCREMENT, action_name TEXT NOT NULL, action_public INTEGER NOT NULL, action_inputs TEXT NOT NULL);"
	case statementsTable:
		return "CREATE TABLE IF NOT EXISTS _action_statements (id INTEGER PRIMARY KEY AUTOINCREMENT, action_id INTEGER NOT NULL, statement BLOB NOT NULL, FOREIGN KEY (action_id) REFERENCES _actions(id));"
	default:
		panic("unknown metadata table")
	}
}

func (m *metadataTable) String() string {
	return string(*m)
}

/*
################################################################
Schema
################################################################
*/
// storeSchema stores the schema in the database.
// It stores the tables and actions in a transaction.
func (d *Dataset) storeSchema(schema *models.Dataset) error {

	err := d.storeTables(schema.Tables)
	if err != nil {
		return fmt.Errorf("failed to store tables: %w", err)
	}

	err = d.storeActions(schema.Actions)
	if err != nil {
		return fmt.Errorf("failed to store actions: %w", err)
	}

	return nil
}

// retrieveSchema retrieves the schema from the database.
func (d *Dataset) retrieveSchema() (*models.Dataset, error) {
	tables, err := d.retrieveTables()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tables: %w", err)
	}

	actions, err := d.retrieveActions()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve actions: %w", err)
	}

	return &models.Dataset{
		Tables:  tables,
		Actions: actions,
		Owner:   d.Owner,
		Name:    d.Name,
	}, nil
}

/*
################################################################
Tables
################################################################
*/
const sqlGetTables = `SELECT id, table_name FROM _tables;`

// retrieveTables retrieves the tables from the database.
func (d *Dataset) retrieveTables() ([]*models.Table, error) {
	tables := make([]*models.Table, 0)

	err := d.conn.Query(sqlGetTables, func(stmt *driver.Statement) error {
		tableId := stmt.GetInt64("id")
		tableName := stmt.GetText("table_name")

		columns, err := d.retrieveColumns(tableId)
		if err != nil {
			return err
		}

		indexes, err := d.retrieveIndexes(tableId)
		if err != nil {
			return err
		}

		tables = append(tables, &models.Table{
			Name:    tableName,
			Columns: columns,
			Indexes: indexes,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return tables, nil
}

const sqlStoreTables = `INSERT INTO _tables (table_name) VALUES ($table_name);`

// storeTables stores the tables in the database and returns the table ids.
func (d *Dataset) storeTables(tables []*models.Table) error {
	for _, table := range tables {
		err := d.conn.ExecuteNamed(sqlStoreTables, map[string]interface{}{
			"$table_name": table.Name,
		})
		if err != nil {
			return err
		}

		tableId, err := d.retrieveTableId(table.Name)
		if err != nil {
			return err
		}

		err = d.storeColumns(tableId, table.Columns)
		if err != nil {
			return err
		}

		err = d.storeIndexes(tableId, table.Indexes)
		if err != nil {
			return err
		}
	}

	return nil
}

const sqlGetTableId = `SELECT id FROM _tables WHERE table_name = $table_id;`

// retrieveTableId retrieves the table id from the database.
func (d *Dataset) retrieveTableId(tableName string) (int64, error) {
	var tableId int64

	err := d.conn.Query(sqlGetTableId, func(stmt *driver.Statement) error {
		tableId = stmt.GetInt64("id")
		return nil
	}, tableName)
	if err != nil {
		return 0, err
	}

	return tableId, nil
}

/*
################################################################
Columns
################################################################
*/
const sqlGetColumns = `SELECT id, column_name, column_type FROM _columns WHERE table_id = $table_id;`

// retrieveColumns retrieves the columns from the database.
func (d *Dataset) retrieveColumns(tableId int64) ([]*models.Column, error) {
	columns := make([]*models.Column, 0)

	err := d.conn.QueryNamed(sqlGetColumns, func(stmt *driver.Statement) error {
		columnId := stmt.GetInt64("id")
		columnName := stmt.GetText("column_name")
		columnType := types.DataType(stmt.GetInt64("column_type"))

		attributes, err := d.retrieveAttributes(columnId)
		if err != nil {
			return err
		}

		columns = append(columns, &models.Column{
			Name:       columnName,
			Type:       columnType,
			Attributes: attributes,
		})

		return nil
	}, map[string]interface{}{
		"$table_id": tableId,
	})
	if err != nil {
		return nil, err
	}

	return columns, nil
}

const sqlStoreColumns = `INSERT INTO _columns (table_id, column_name, column_type) VALUES ($table_id, $column_name, $column_type);`

// storeColumns stores the columns in the database and returns the column ids.
func (d *Dataset) storeColumns(tableId int64, columns []*models.Column) error {
	for _, column := range columns {
		err := d.conn.ExecuteNamed(sqlStoreColumns, map[string]interface{}{
			"$table_id":    tableId,
			"$column_name": column.Name,
			"$column_type": column.Type,
		})
		if err != nil {
			return err
		}

		columnId, err := d.retrieveColumnId(tableId, column.Name)
		if err != nil {
			return err
		}

		err = d.storeAttributes(columnId, column.Attributes)
		if err != nil {
			return err
		}
	}

	return nil
}

const sqlGetColumnId = `SELECT id FROM _columns WHERE table_id = $table_id AND column_name = $column_name;`

// retrieveColumnId retrieves the column id from the database.
func (d *Dataset) retrieveColumnId(tableId int64, columnName string) (int64, error) {
	var columnId int64

	err := d.conn.QueryNamed(sqlGetColumnId, func(stmt *driver.Statement) error {
		columnId = stmt.GetInt64("id")
		return nil
	}, map[string]interface{}{
		"$table_id":    tableId,
		"$column_name": columnName,
	})
	if err != nil {
		return 0, err
	}

	return columnId, nil
}

/*
################################################################
Attributes
################################################################
*/
const sqlGetAttributes = `SELECT attribute_type, attribute_value FROM _attributes WHERE column_id = $column_id;`

// retrieveAttributes retrieves the attributes from the database.
func (d *Dataset) retrieveAttributes(columnId int64) ([]*models.Attribute, error) {
	attributes := make([]*models.Attribute, 0)

	err := d.conn.QueryNamed(sqlGetAttributes, func(stmt *driver.Statement) error {
		attributeType := types.AttributeType(stmt.GetInt64("attribute_type"))
		attributeValue, _ := stmt.GetBytes("attribute_value")

		attributes = append(attributes, &models.Attribute{
			Type:  attributeType,
			Value: attributeValue,
		})

		return nil
	}, map[string]interface{}{
		"$column_id": columnId,
	})
	if err != nil {
		return nil, err
	}

	return attributes, nil
}

const sqlStoreAttributes = `INSERT INTO _attributes (column_id, attribute_type, attribute_value) VALUES ($column_id, $attribute_type, $attribute_value);`

// storeAttributes stores the attributes in the database.
func (d *Dataset) storeAttributes(columnId int64, attributes []*models.Attribute) error {
	for _, attribute := range attributes {
		err := d.conn.ExecuteNamed(sqlStoreAttributes, map[string]interface{}{
			"$column_id":       columnId,
			"$attribute_type":  attribute.Type,
			"$attribute_value": attribute.Value,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

/*
################################################################
Indexes
################################################################
*/
const sqlGetIndexes = `SELECT index_name, index_type, columns FROM _indexes WHERE table_id = $table_id;`

// retrieveIndexes retrieves the indexes from the database.
func (d *Dataset) retrieveIndexes(tableId int64) ([]*models.Index, error) {
	indexes := make([]*models.Index, 0)

	err := d.conn.QueryNamed(sqlGetIndexes, func(stmt *driver.Statement) error {
		indexName := stmt.GetText("index_name")
		indexType := types.IndexType(stmt.GetInt64("index_type"))
		columns := getDelimited(stmt.GetText("columns"))

		indexes = append(indexes, &models.Index{
			Name:    indexName,
			Type:    indexType,
			Columns: columns,
		})

		return nil
	}, map[string]interface{}{
		"$table_id": tableId,
	})
	if err != nil {
		return nil, err
	}

	return indexes, nil
}

const sqlStoreIndexes = `INSERT INTO _indexes (table_id, index_name, index_type, columns) VALUES ($table_id, $index_name, $index_type, $columns);`

// storeIndexes stores the indexes in the database.
func (d *Dataset) storeIndexes(tableId int64, indexes []*models.Index) error {
	for _, index := range indexes {
		err := d.conn.ExecuteNamed(sqlStoreIndexes, map[string]interface{}{
			"$table_id":   tableId,
			"$index_name": index.Name,
			"$index_type": index.Type,
			"$columns":    strings.Join(index.Columns, ","),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

/*
################################################################
Actions
################################################################
*/
const sqlGetActions = `SELECT id, action_name, action_public, action_inputs FROM _actions;`

// retrieveActions retrieves the actions from the database.
func (d *Dataset) retrieveActions() ([]*models.Action, error) {
	actions := make([]*models.Action, 0)

	err := d.conn.Query(sqlGetActions, func(stmt *driver.Statement) error {
		actionId := stmt.GetInt64("id")
		actionName := stmt.GetText("action_name")
		actionPublic := intToBool(int(stmt.GetInt64("action_public")))
		inputs := stmt.GetText("action_inputs")

		var actionInputs []string
		if inputs != "" {
			actionInputs = getDelimited(stmt.GetText("action_inputs"))
		}

		statements, err := d.retrieveActionStatements(actionId)
		if err != nil {
			return err
		}

		actions = append(actions, &models.Action{
			Name:       actionName,
			Public:     actionPublic,
			Inputs:     actionInputs,
			Statements: statements,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return actions, nil
}

const sqlStoreActions = `INSERT INTO _actions (action_name, action_public, action_inputs) VALUES ($action_name, $action_public, $action_inputs);`

// storeActions stores the actions in the database.
func (d *Dataset) storeActions(actions []*models.Action) error {
	for _, action := range actions {
		err := d.conn.ExecuteNamed(sqlStoreActions, map[string]interface{}{
			"$action_name":   action.Name,
			"$action_public": boolToInt(action.Public),
			"$action_inputs": strings.Join(action.Inputs, ","),
		})
		if err != nil {
			return err
		}

		actionId, err := d.getActionId(action.Name)
		if err != nil {
			return err
		}

		err = d.storeActionStatements(actionId, action.Statements)
		if err != nil {
			return err
		}
	}

	return nil
}

const sqlGetActionId = `SELECT id FROM _actions WHERE action_name = $action_name;`

// getActionId retrieves the action id from the database.
func (d *Dataset) getActionId(actionName string) (int64, error) {
	var actionId int64

	err := d.conn.QueryNamed(sqlGetActionId, func(stmt *driver.Statement) error {
		actionId = stmt.GetInt64("id")
		return nil
	}, map[string]interface{}{
		"$action_name": actionName,
	})

	return actionId, err
}

/*
################################################################
Statements
################################################################
*/
const sqlGetActionStatements = `SELECT statement FROM _action_statements WHERE action_id = $action_id;`

// retrieveActionStatements retrieves the action statements from the database.
func (d *Dataset) retrieveActionStatements(actionId int64) ([]string, error) {
	statements := make([]string, 0)

	err := d.conn.QueryNamed(sqlGetActionStatements, func(stmt *driver.Statement) error {
		statement := stmt.GetText("statement")
		statements = append(statements, statement)
		return nil
	}, map[string]interface{}{
		"$action_id": actionId,
	})
	if err != nil {
		return nil, err
	}

	return statements, nil
}

const sqlStoreActionStatements = `INSERT INTO _action_statements (action_id, statement) VALUES ($action_id, $statement);`

// storeActionStatements stores the action statements in the database.
func (d *Dataset) storeActionStatements(actionId int64, statements []string) error {
	for _, statement := range statements {
		err := d.conn.ExecuteNamed(sqlStoreActionStatements, map[string]interface{}{
			"$action_id": actionId,
			"$statement": serializeStmt(statement),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func serializeStmt(stmt string) []byte {
	return []byte(stmt)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i == 1
}

func getDelimited(s string) []string {
	var result []string
	if s != "" {
		result = strings.Split(s, ",")
	}

	return result
}
