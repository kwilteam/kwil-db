package datasets

import (
	"fmt"
	sqlitegenerator "github.com/kwilteam/kwil-db/pkg/engine/datasets/sqlite-generator"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/engine/models/validation"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql/driver"
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
	path         string
	log          log.Logger
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
func OpenDataset(owner, name string, opts ...DatasetConnectionOpts) (*Dataset, error) {
	dbid := models.GenerateSchemaId(owner, name)

	ds := &Dataset{
		conn:         nil,
		Owner:        owner,
		Name:         name,
		DBID:         dbid,
		Tables:       make(map[string]*models.Table),
		actions:      make(map[string]*PreparedAction),
		path:         "",
		log:          log.NewNoOp(),
		readOnlyConn: nil,
	}

	for _, opt := range opts {
		opt(ds)
	}

	var err error
	ds.conn, err = driver.OpenConn(dbid,
		ds.driverOpts()...,
	)
	if err != nil {
		return nil, err
	}

	ds.readOnlyConn, err = ds.conn.CopyReadOnly()
	if err != nil {
		return nil, err
	}

	err = ds.conn.AcquireLock()
	if err != nil {
		return nil, err
	}

	err = ds.init()
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (ds *Dataset) driverOpts() []driver.ConnOpt {
	opts := []driver.ConnOpt{
		driver.WithInjectableVars(injectableVars),
		driver.WithLogger(ds.log),
	}

	if ds.path != "" {
		opts = append(opts, driver.WithPath(ds.path))
	}

	return opts
}

// Close closes the dataset and releases the lock.
func (d *Dataset) Close() error {
	if d.readOnlyConn != nil {
		d.readOnlyConn.ReleaseLock()
		d.readOnlyConn.Close()
	}
	d.conn.ReleaseLock()
	return d.conn.Close()
}

// init initializes the underlying sqlite database.
// it will also load the schema and actions from disk if it exists.
func (d *Dataset) init() error {
	for _, tbl := range metadataTables {
		exists, err := d.conn.TableExists(tbl.String())
		if err != nil {
			return fmt.Errorf("error checking if table exists: %w", err)
		}

		if !exists {
			err := d.conn.Execute(tbl.initStmt())
			if err != nil {
				return fmt.Errorf("error initializing dataset metadata tables: %w", err)
			}
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

	sp, err := d.conn.Begin()
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
			return fmt.Errorf("error preparing action on schema apply %s: %w", action.Name, err)
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
			return fmt.Errorf("error preparing action on schema load %s: %w", action.Name, err)
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
