package dataset

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
)

// DatasetContext is a context for a dataset.
// Once provided, it should not be modified.
type DatasetContext struct {
	// IdFunc is a function that generates a dbid from an owner and name.
	IdFunc func(owner, name string) string

	// Name is the name of the dataset.
	Name string
	// Owner is the owner of the dataset.
	Owner string
}

// A database is a single deployed instance of kwil-db.
// It contains a SQLite file
type Dataset struct {
	Ctx     *DatasetContext
	db      DB
	actions map[string]*Action
	tables  map[string]*dto.Table
}

// NewDataset creates a new dataset.
func NewDataset(ctx *DatasetContext, db DB) *Dataset {
	return &Dataset{
		Ctx:     ctx,
		db:      db,
		actions: make(map[string]*Action),
		tables:  make(map[string]*dto.Table),
	}
}

// CreateAction creates a new action and prepares it for use.
func (d *Dataset) CreateAction(a *dto.Action) (*Action, error) {
	if d.actions[strings.ToLower(a.Name)] != nil {
		return nil, fmt.Errorf(`action "%s" already exists`, a.Name)
	}

	action := &Action{
		Action:  a,
		stmts:   make([]Statement, len(a.Statements)),
		dataset: d,
	}

	for i, stmt := range a.Statements {
		stmt, err := d.db.Prepare(stmt)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %w", err)
		}

		action.stmts[i] = stmt

	}

	d.actions[strings.ToLower(action.Name)] = action

	return action, nil
}

func (d *Dataset) Close() error {
	for _, action := range d.actions {
		err := action.Close()
		if err != nil {
			return err
		}
	}

	return d.db.Close()
}
