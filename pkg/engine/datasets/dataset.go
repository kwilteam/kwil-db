package datasets

import (
	"fmt"
	sqlitegenerator "kwil/pkg/engine/datasets/sqlite-generator"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/models/validation"
	"kwil/pkg/sql/driver"
)

var (
	injectableVars = []*driver.InjectableVar{
		{
			Name:       "@caller",
			DefaultVal: "0x0000000000000000000000000000000000000000",
		},
	}
)

type Dataset struct {
	conn         *driver.Connection
	Owner        string
	Name         string
	DBID         string
	Tables       map[string]*models.Table
	actions      map[string]*PreparedAction
	readOnlyConn *driver.Connection
}

// prepareAction prepares an action for execution at a later time.
func (d *Dataset) prepareAction(action *models.Action) error {
	acc, err := NewPreparedAction(d.conn, action, d.Tables)
	if err != nil {
		return err
	}

	d.actions[action.Name] = acc
	return nil
}

// OpenDataset opens a dataset with the given owner, name, and path.
// if the dataset does not exist, it will be created.
func OpenDataset(owner, name, path string) (*Dataset, error) {
	dbid := models.GenerateSchemaId(owner, name)

	conn, err := driver.OpenConn(dbid,
		driver.WithPath(path),
		driver.WithInjectableVars(injectableVars),
	)
	if err != nil {
		return nil, err
	}

	readOnlyCon, err := conn.CopyReadOnly()
	if err != nil {
		return nil, err
	}

	err = conn.AcquireLock()
	if err != nil {
		return nil, err
	}

	d := &Dataset{
		conn:         conn,
		Owner:        owner,
		Name:         name,
		DBID:         dbid,
		Tables:       make(map[string]*models.Table),
		actions:      make(map[string]*PreparedAction),
		readOnlyConn: readOnlyCon,
	}

	err = d.init()
	if err != nil {
		return nil, err
	}

	return d, nil
}

// Close closes the dataset and releases the lock.
func (d *Dataset) Close() error {
	d.conn.ReleaseLock()
	return d.conn.Close()
}

// init initializes the underlying sqlite database.
// it will also load the schema and actions from disk if it exists.
func (d *Dataset) init() error {
	stmts := getTableInits()
	for _, stmt := range stmts {
		err := d.conn.Execute(stmt)
		if err != nil {
			return fmt.Errorf("error initializing dataset metadata tables: %w", err)
		}
	}

	return d.loadSchema()
}

// ApplySchema applies the given schema to the dataset.
func (d *Dataset) ApplySchema(schema *models.Dataset) (err error) {
	if len(d.actions) > 0 || len(d.Tables) > 0 {
		return fmt.Errorf("schema already exists for dataset %s", d.DBID)
	}

	err = validation.Validate(schema)
	if err != nil {
		return fmt.Errorf("error on schema: %w", err)
	}

	sp, err := d.conn.Savepoint()
	if err != nil {
		return fmt.Errorf("error creating savepoint: %w", err)
	}
	defer sp.Rollback()

	for _, table := range schema.Tables {
		err = d.applyTable(table)
		if err != nil {
			return fmt.Errorf("error applying table %s: %w", table.Name, err)
		}
	}

	for _, action := range schema.Actions {
		err := d.prepareAction(action)
		if err != nil {
			return fmt.Errorf("error preparing action %s: %w", action.Name, err)
		}
	}

	for _, table := range schema.Tables {
		d.Tables[table.Name] = table
	}

	err = d.storeSchema(schema)
	if err != nil {
		return fmt.Errorf("error storing schema: %w", err)
	}

	return sp.Commit()
}

// loadSchema loads the schema from disk.
func (d *Dataset) loadSchema() error {
	schema, err := d.retrieveSchema()
	if err != nil {
		return fmt.Errorf("error retrieving schema: %w", err)
	}

	for _, action := range schema.Actions {
		err := d.prepareAction(action)
		if err != nil {
			return fmt.Errorf("error preparing action %s: %w", action.Name, err)
		}
	}

	for _, table := range schema.Tables {
		d.Tables[table.Name] = table
	}

	return nil
}

// applyTable applies the given table to the dataset.
func (d *Dataset) applyTable(table *models.Table) error {
	ddl, err := sqlitegenerator.GenerateDDL(table)
	if err != nil {
		return fmt.Errorf("error generating DDL for table %s: %w", table.Name, err)
	}

	for _, stmt := range ddl {
		err := d.conn.Execute(stmt)
		if err != nil {
			return fmt.Errorf("error applying table %s: %w", table.Name, err)
		}
	}

	return nil
}

// Wipe wipes the dataset.
func (d *Dataset) Wipe() error {
	err := d.conn.DisableForeignKeys()
	if err != nil {
		return fmt.Errorf("error disabling foreign keys: %w", err)
	}

	tables := make([]string, 0)
	err = d.conn.Query("SELECT name FROM sqlite_master WHERE type='table' AND name not like 'sqlite_sequence';", func(stmt *driver.Statement) error {
		tables = append(tables, stmt.GetText("name"))
		return nil
	})
	if err != nil {
		return fmt.Errorf("error retrieving tables: %w", err)
	}

	for _, table := range tables {
		err := d.conn.Execute(fmt.Sprintf("DROP TABLE %s;", table))
		if err != nil {
			return fmt.Errorf("error dropping table %s: %w", table, err)
		}
	}

	err = d.conn.EnableForeignKeys()
	if err != nil {
		return fmt.Errorf("error enabling foreign keys: %w", err)
	}

	return nil
}

// Clear wipes a database and reinitializes it.
func (d *Dataset) Clear() error {
	err := d.Wipe()
	if err != nil {
		return fmt.Errorf("error wiping dataset: %w", err)
	}

	d.Tables = make(map[string]*models.Table)
	d.actions = make(map[string]*PreparedAction)

	return d.init()
}

func (d *Dataset) GetSchema() *models.Dataset {
	tbls := make([]*models.Table, 0)
	for _, tbl := range d.Tables {
		tbls = append(tbls, tbl)
	}

	acts := make([]*models.Action, 0)
	for _, act := range d.actions {
		acts = append(acts, act.GetAction())
	}

	return &models.Dataset{
		Owner:   d.Owner,
		Name:    d.Name,
		Tables:  tbls,
		Actions: acts,
	}
}
