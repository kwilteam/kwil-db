package dataset

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/engine2/sqldb"
	"github.com/kwilteam/kwil-db/pkg/engine2/utils"
)

// DatasetContext is a context for a dataset.
// Once provided, it should not be modified.
type DatasetContext struct {
	// Name is the name of the dataset.
	Name string
	// Owner is the owner of the dataset.
	Owner string
}

// A database is a single deployed instance of kwil-db.
// It contains a SQLite file
type Dataset struct {
	Ctx     *DatasetContext
	db      sqldb.DB
	actions map[string]*preparedAction
	tables  map[string]*dto.Table
}

// NewDataset creates a new dataset.
func NewDataset(ctx context.Context, dsCtx *DatasetContext, db sqldb.DB) (*Dataset, error) {
	ds := &Dataset{
		Ctx:     dsCtx,
		db:      db,
		actions: make(map[string]*preparedAction),
		tables:  make(map[string]*dto.Table),
	}

	tables, err := db.ListTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	for _, table := range tables {
		ds.tables[strings.ToLower(table.Name)] = table
	}

	actions, err := db.ListActions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list actions: %w", err)
	}

	for _, action := range actions {
		prepAction, err := ds.prepareAction(action)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare action: %w", err)
		}

		ds.actions[strings.ToLower(action.Name)] = prepAction
	}

	return ds, nil
}

// CreateAction creates a new action and prepares it for use.
func (d *Dataset) CreateAction(ctx context.Context, actionDefinition *dto.Action) error {
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
func (d *Dataset) GetAction(name string) *dto.Action {
	return d.actions[strings.ToLower(name)].Action
}

// ListActions returns a list of actions.
func (d *Dataset) ListActions() []*dto.Action {
	actions := make([]*dto.Action, 0, len(d.actions))
	for _, action := range d.actions {
		actions = append(actions, action.Action)
	}

	return actions
}

// CreateTable creates a new table and prepares it for use.
func (d *Dataset) CreateTable(ctx context.Context, table *dto.Table) error {
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
func (d *Dataset) GetTable(name string) *dto.Table {
	return d.tables[strings.ToLower(name)]
}

// ListTables returns a list of tables.
func (d *Dataset) ListTables() []*dto.Table {
	tables := make([]*dto.Table, 0, len(d.tables))
	for _, table := range d.tables {
		tables = append(tables, table)
	}

	return tables
}

// Close closes the dataset.
func (d *Dataset) Close() error {
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
	return utils.GenerateDBID(d.Ctx.Name, d.Ctx.Owner)
}

// Execute executes an action.
// It will execute as many times as there are inputs, and will return the last result.
func (d *Dataset) Execute(txCtx *dto.TxContext, inputs []map[string]any) (dto.Result, error) {
	action, ok := d.actions[strings.ToLower(txCtx.Action)]
	if !ok {
		return nil, fmt.Errorf(`action "%s" does not exist`, txCtx.Action)
	}

	return action.BatchExecute(txCtx, inputs)
}

func (d *Dataset) Savepoint() (sqldb.Savepoint, error) {
	return d.db.Savepoint()
}
