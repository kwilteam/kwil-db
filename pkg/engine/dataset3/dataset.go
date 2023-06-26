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
	Procedures() []*types.Procedure
	Tables(ctx context.Context) []*types.Table
	Delete() error
	Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Execute(ctx context.Context, procedure string, args []map[string]any, opts *TxOpts) ([]map[string]any, error)
}

// A dataset is a deployed Kwil database with an underlying data store and engine.
type Dataset struct {
	metadata *Metadata
	db       Datastore
	engine   Engine
}

// OpenDataset opens a new dataset and loads the metadata from the database
func OpenDataset(ctx context.Context, ds Datastore) (*Dataset, error) {
	procedures, err := getProcedureMap(ctx, ds)
	if err != nil {
		return nil, err
	}

	return &Dataset{
		metadata: &Metadata{
			Procedures: procedures,
		},
		db: ds,
	}, nil
}

// getProcedureMap returns a map of procedure names to procedures.
func getProcedureMap(ctx context.Context, ds Datastore) (map[string]*types.Procedure, error) {
	procs, err := ds.ListProcedures(ctx)
	if err != nil {
		return nil, err
	}

	procMap := make(map[string]*types.Procedure)
	for _, proc := range procs {
		procMap[proc.Name] = proc
	}

	return procMap, nil
}

// Query executes a ad-hoc, read-only query.
func (d *Dataset) Query(ctx context.Context, stmt string, args map[string]any) (io.Reader, error) {
	return d.db.Query(ctx, stmt, args)
}

// Procedures returns the procedures in the dataset.
func (d *Dataset) Procedures() []*types.Procedure {
	var procs []*types.Procedure
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
