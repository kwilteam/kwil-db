package dataset3

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

// TODO: This is a stub. Delete it.
type IDataset interface {
	Close() error
	Procedures() []*Procedure
	Tables(ctx context.Context) []*types.Table
	Delete() error
	Query(ctx context.Context, stmt string, args map[string]any) (io.Reader, error)
	Execute(ctx context.Context, stmt string, args map[string]any, opts *TxOpts) (io.Reader, error)
}

// A dataset is a deployed Kwil database with an underlying data store and engine.
type Dataset struct {
	dbid string

	metadata *Metadata
	db       Datastore
	engine   Engine
}

// OpenDataset wraps the database with a Dataset.
// TODO: Should this be renamed?
func OpenDataset(ctx context.Context, ds Dataset) (*Dataset, error) {
	return nil, nil
}

// Query executes a ad-hoc, read-only query.
func (d *Dataset) Query(ctx context.Context, stmt string, args map[string]any) (io.Reader, error) {
	return d.db.Query(ctx, stmt, args)
}

// Procedures returns the procedures in the dataset.
func (d *Dataset) Procedures() []*Procedure {
	var procs []*Procedure
	for _, procedure := range d.metadata.Procedures {
		procs = append(procs, procedure)
	}

	return procs
}

// Tables returns the tables in the dataset.
func (d *Dataset) Tables(ctx context.Context) ([]*types.Table, error) {
	return d.db.ListTables(ctx)
}

// Close closes the dataset.
func (d *Dataset) Close() error {
	var errs []string

	err := d.engine.Close()
	if err != nil {
		errs = append(errs, err.Error())
	}

	err = d.db.Close()
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing dataset: %s", strings.Join(errs, ", "))
	}

	return nil
}

// Delete deletes the dataset.
func (d *Dataset) Delete() error {
	var errs []string

	err := d.engine.Close()
	if err != nil {
		errs = append(errs, err.Error())
	}

	err = d.db.Delete()
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors deleting dataset: %s", strings.Join(errs, ", "))
	}

	return nil
}
