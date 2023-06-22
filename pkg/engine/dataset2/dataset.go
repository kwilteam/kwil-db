package dataset2

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
	mu         sync.RWMutex
	name       string
	owner      string
	db         Datastore
	tables     map[string]*dto.Table
	procedures map[string]*StoredProcedure
	onLoad     []*InstructionExecution
	onDeploy   []*InstructionExecution

	cache *Cache
}

// OpenDataset initializes a Dataset struct and loads tables, actions, and extensions.
// It accepts a context, database interface and a variadic number of options to customize the Dataset.
// It returns a pointer to the Dataset and an error if any operation fails.
func OpenDataset(ctx context.Context, db Datastore, opts ...OpenOpt) (*Dataset, error) {
	ds := &Dataset{
		mu:     sync.RWMutex{},
		db:     db,
		tables: make(map[string]*dto.Table),
		cache:  newCache(),
	}

	for _, opt := range opts {
		opt(ds)
	}

	err := ds.loadTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load tables: %w", err)
	}

	err = ds.loadProcedures(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load actions: %w", err)
	}

	return ds, nil
}

func (d *Dataset) evaluateOnLoad(ctx context.Context) error {
	execCtx := newExecutionContext(ctx, "_load", WithCaller(d.owner), WithDataset(d.name))

	evaluateInstructions(execCtx, d, d.onLoad, execCtx)
}

// loadTables loads the tables from the database and stores them in the Dataset's 'tables' map.
// It accepts a context and returns an error if the operation fails.
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

// loadProcedures loads the actions from the database, prepares them, and stores them in the Dataset's 'actions' map.
// It accepts a context and returns an error if the operation fails.
// TODO: change this to new procedures
func (d *Dataset) loadProcedures(ctx context.Context) error {
	procs, err := d.db.ListProcedures(ctx)
	if err != nil {
		return fmt.Errorf("failed to list actions: %w", err)
	}

	for _, action := range procs {
		prepAction, err := d.prepareProcedure(action)
		if err != nil {
			return fmt.Errorf("failed to prepare action: %w", err)
		}

		d.cache.procedures[strings.ToLower(action.Name)] = prepAction
	}

	return nil
}

// CreateAction prepares and stores a new action in the Dataset's 'actions' map.
// It accepts a context and an Action struct and returns an error if the operation fails.
// TODO: delete, will be replaced with CreateProcedure
func (d *Dataset) CreateAction(ctx context.Context, actionDefinition *dto.Action) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cache.procedures[strings.ToLower(actionDefinition.Name)] != nil {
		return fmt.Errorf(`action "%s" already exists`, actionDefinition.Name)
	}

	newProcedure, err := d.prepareProcedure(actionDefinition)
	if err != nil {
		return fmt.Errorf("failed to prepare action: %w", err)
	}

	err = d.db.StoreProcedure(ctx, actionDefinition)
	if err != nil {
		return fmt.Errorf("failed to store action: %w", err)
	}

	d.cache.procedures[strings.ToLower(newProcedure.Name)] = newProcedure

	return nil
}

/*
func (d *Dataset) CreateProcedure(ctx context.Context, procedure *Procedure) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	lowerName := strings.ToLower(procedure.Name)
	if _, ok := d.cache.procedures[lowerName]; ok {
		return fmt.Errorf(`procedure "%s" already exists`, procedure.Name)
	}

}*/

// TODO: implement
func (d *Dataset) prepareProcedure(action *dto.Action) (*StoredProcedure, error) {
	return nil, nil
	/*
		newAction := &procedure{
			Action:     action,
			operations: make([]operation, len(action.Statements)),
			dataset:    d,
		}

		for i, statement := range action.Statements {
			stmt, err := d.db.Prepare(statement)
			if err != nil {
				return nil, fmt.Errorf("failed to prepare statement: %w", err)
			}

			newAction.operations[i] = stmt
		}

		return newAction, nil
	*/
}

// GetAction retrieves an action from the Dataset's 'actions' map by its name.
// It returns the Action struct or nil if the action does not exist.
func (d *Dataset) GetAction(name string) (*dto.Action, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	preparedProcedure, ok := d.cache.procedures[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf(`action "%s" does not exist`, name)
	}

	return preparedProcedure.Action, nil
}

// ListProcedures returns a list of procedures.
func (d *Dataset) ListProcedures() []*dto.Action {
	d.mu.RLock()
	defer d.mu.RUnlock()

	actions := make([]*dto.Action, 0, len(d.cache.procedures))
	for _, action := range d.cache.procedures {
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

// Close closes the dataset, freeing up resources.
// It closes all actions and the database connection, returning an error if any operation fails.
func (d *Dataset) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, action := range d.cache.procedures {
		err := action.close()
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

// Execute executes a procedure atomically.
// It accepts a context, the procedure name, a double slice of inputs, and options.
// It returns the result of the last execution and an error if any operation fails.
func (d *Dataset) Execute(ctx context.Context, procedureName string, inputs [][]any, opts ...ExecutionOpt) (dto.Result, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	execCtx := newExecutionContext(ctx, procedureName, opts...)

	procedure, ok := d.cache.procedures[procedureName]
	if !ok {
		return nil, fmt.Errorf(`action "%s" does not exist`, procedureName)
	}

	if len(inputs) == 0 {
		inputs = append(inputs, []any{}) // this will cause it to execute once with no inputs
	}

	for _, input := range inputs {
		err := procedure.Execute(execCtx, input)
		if err != nil {
			return nil, err
		}
	}

	return execCtx.lastDmlResult, nil
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

	for _, action := range d.cache.procedures {
		err := action.close()
		if err != nil {
			return err
		}
	}

	return d.db.Delete()
}

// Query performs a read-only query on the dataset.
// It accepts a context, a query string, and a map of arguments. It returns the query result and an error if the operation fails.
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
