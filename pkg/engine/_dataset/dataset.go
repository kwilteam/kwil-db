package dataset

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/sqldb"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
)

// A database is a single deployed instance of kwil-db.
// It contains a SQLite file
type Dataset struct {
	mu                    sync.RWMutex
	name                  string
	owner                 string
	db                    sqldb.DB
	actions               map[string]*preparedAction
	tables                map[string]*dto.Table
	extensions            map[string]InitializedExtension
	extensionInitializers map[string]Initializer
}

func OpenDataset(ctx context.Context, db sqldb.DB, opts ...OpenOpt) (*Dataset, error) {
	ds := &Dataset{
		mu:                    sync.RWMutex{},
		db:                    db,
		actions:               make(map[string]*preparedAction),
		tables:                make(map[string]*dto.Table),
		extensions:            make(map[string]InitializedExtension),
		extensionInitializers: make(map[string]Initializer),
	}

	for _, opt := range opts {
		opt(ds)
	}

	err := ds.loadTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load tables: %w", err)
	}

	err = ds.loadActions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load actions: %w", err)
	}

	err = ds.loadExtensions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load extensions: %w", err)
	}

	return ds, nil
}

func (d *Dataset) loadTables(ctx context.Context) error {
	tables, err := d.db.ListTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	for _, table := range tables {
		d.tables[strings.ToLower(table.Name)] = table
	}

	return nil
}

func (d *Dataset) loadActions(ctx context.Context) error {
	actions, err := d.db.ListActions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list actions: %w", err)
	}

	for _, action := range actions {
		prepAction, err := d.prepareAction(action)
		if err != nil {
			return fmt.Errorf("failed to prepare action: %w", err)
		}

		d.actions[strings.ToLower(action.Name)] = prepAction
	}

	return nil
}

func (d *Dataset) loadExtensions(ctx context.Context) error {
	exts, err := d.db.GetExtensions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get extensions: %w", err)
	}

	for _, ext := range exts {
		initializer, ok := d.extensionInitializers[ext.Name]
		if !ok {
			return fmt.Errorf("schema requires extension %s, but it is not registered on this node", ext.Name)
		}

		initializedExt, err := initializer.Initialize(ctx, ext.Metadata)
		if err != nil {
			return fmt.Errorf("failed to initialize extension %s: %w", ext.Name, err)
		}

		d.extensions[ext.Name] = initializedExt
	}

	return nil
}

// CreateAction creates a new action and prepares it for use.
func (d *Dataset) CreateAction(ctx context.Context, actionDefinition *dto.Action) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.actions[strings.ToLower(actionDefinition.Name)] != nil {
		return fmt.Errorf(`action "%s" already exists`, actionDefinition.Name)
	}

	newAction, err := d.prepareAction(actionDefinition)
	if err != nil {
		return fmt.Errorf("failed to prepare action: %w", err)
	}

	err = d.db.StoreAction(ctx, actionDefinition)
	if err != nil {
		return fmt.Errorf("failed to store action: %w", err)
	}

	d.actions[strings.ToLower(newAction.Name)] = newAction

	return nil
}

func (d *Dataset) prepareAction(action *dto.Action) (*preparedAction, error) {
	newAction := &preparedAction{
		Action:  action,
		stmts:   make([]sqldb.Statement, len(action.Statements)),
		dataset: d,
	}

	for i, statement := range action.Statements {
		stmt, err := d.db.Prepare(statement)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %w", err)
		}

		newAction.stmts[i] = stmt
	}

	return newAction, nil
}

// GetAction returns an action by name.
func (d *Dataset) GetAction(name string) (*dto.Action, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	preparedActiom, ok := d.actions[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf(`action "%s" does not exist`, name)
	}

	return preparedActiom.Action, nil
}

// ListActions returns a list of actions.
func (d *Dataset) ListActions() []*dto.Action {
	d.mu.RLock()
	defer d.mu.RUnlock()

	actions := make([]*dto.Action, 0, len(d.actions))
	for _, action := range d.actions {
		actions = append(actions, action.Action)
	}

	return actions
}

// CreateTable creates a new table and prepares it for use.
func (d *Dataset) CreateTable(ctx context.Context, table *dto.Table) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.tables[strings.ToLower(table.Name)] != nil {
		return fmt.Errorf(`table "%s" already exists`, table.Name)
	}

	err := d.db.CreateTable(ctx, table)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	d.tables[strings.ToLower(table.Name)] = table

	return nil
}

// GetTable returns a table by name.
func (d *Dataset) GetTable(name string) (*dto.Table, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	tbl, ok := d.tables[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf(`table "%s" does not exist`, name)
	}

	return tbl, nil
}

// ListTables returns a list of tables.
func (d *Dataset) ListTables() []*dto.Table {
	d.mu.RLock()
	defer d.mu.RUnlock()

	tables := make([]*dto.Table, 0, len(d.tables))
	for _, table := range d.tables {
		tables = append(tables, table)
	}

	return tables
}

// Close closes the dataset.
func (d *Dataset) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, action := range d.actions {
		err := action.Close()
		if err != nil {
			return err
		}
	}

	return d.db.Close()
}

// Id returns the id of the dataset.
func (d *Dataset) Id() string {
	return utils.GenerateDBID(d.name, d.owner)
}

// Execute executes an action.
// It will execute as many times as there are inputs, and will return the last result.
func (d *Dataset) Execute(txCtx *dto.TxContext, inputs []map[string]any) (dto.Result, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	action, ok := d.actions[strings.ToLower(txCtx.Action)]
	if !ok {
		return nil, fmt.Errorf(`action "%s" does not exist`, txCtx.Action)
	}

	if len(inputs) == 0 {
		return action.Execute(txCtx, nil)
	}

	return action.BatchExecute(txCtx, inputs)
}

// Savepoint creates a new savepoint.
func (d *Dataset) Savepoint() (sqldb.Savepoint, error) {
	return d.db.Savepoint()
}

// Delete deletes the dataset.
func (d *Dataset) Delete(txCtx *dto.TxContext) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if txCtx.Caller != d.owner {
		return fmt.Errorf("caller does not have permission to delete dataset")
	}

	for _, action := range d.actions {
		err := action.Close()
		if err != nil {
			return err
		}
	}

	return d.db.Delete()
}

// Query executes a query and returns the result.
// It is a read-only operation.
func (d *Dataset) Query(ctx context.Context, stmt string, args map[string]any) (dto.Result, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.db.Query(ctx, stmt, args)
}

// Owner returns the owner of the dataset.
func (d *Dataset) Owner() string {
	return d.owner
}

// Name returns the name of the dataset.
func (d *Dataset) Name() string {
	return d.name
}
