package dataset

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/pkg/engine/eng"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

// A dataset is a deployed Kwil database with an underlying data store and engine.
type Dataset struct {
	metadata *Metadata
	db       Datastore
	engine   Engine
	options  *engineOptions
}

// OpenDataset opens a new dataset and loads the metadata from the database
func OpenDataset(ctx context.Context, ds Datastore, opts ...OpenOpt) (*Dataset, error) {
	openOptions := &engineOptions{
		initializers: make(map[string]Initializer),
		log:          log.NewNoOp(),
	}
	for _, opt := range opts {
		opt(openOptions)
	}

	dataset := &Dataset{
		db:      ds,
		options: openOptions,
	}

	engineOpts, err := dataset.getEngineOpts(ctx, openOptions.initializers)
	if err != nil {
		return nil, err
	}

	engine, err := eng.NewEngine(ctx, datastoreWrapper{ds}, engineOpts)
	if err != nil {
		return nil, err
	}

	procedures, err := ds.ListProcedures(ctx)
	if err != nil {
		return nil, err
	}

	dataset.engine = engine
	dataset.metadata = newMetadata(procedures)

	return dataset, nil
}

func (d *Dataset) execConstructor(ctx context.Context, opts *TxOpts) error {
	for _, procedure := range d.metadata.Procedures {
		if strings.EqualFold(procedure.Name, constructorName) {
			_, err := d.executeOnce(ctx, procedure, make(map[string]any), opts)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}

const constructorName = "constructor"

// Execute executes a procedure.
func (d *Dataset) Execute(ctx context.Context, action string, args []map[string]any, opts *TxOpts) ([]map[string]any, error) {
	if strings.EqualFold(action, constructorName) {
		return nil, fmt.Errorf("cannot execute constructor")
	}

	proc, ok := d.metadata.Procedures[action]
	if !ok {
		return nil, fmt.Errorf("procedure %s does not exist", action)
	}

	savepoint, err := d.db.Savepoint()
	if err != nil {
		return nil, err
	}
	defer savepoint.Rollback()

	if len(args) == 0 { // if no args, add an empty arg map so we can execute once
		args = append(args, make(map[string]any))
	}

	var result []map[string]any
	for _, arg := range args {
		result, err = d.executeOnce(ctx, proc, arg, opts)
		if err != nil {
			return nil, err
		}
	}

	err = savepoint.Commit()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Dataset) executeOnce(ctx context.Context, proc *types.Procedure, args map[string]any, opts *TxOpts) ([]map[string]any, error) {
	var argArr []any
	for _, arg := range proc.Args {
		val, ok := args[arg]
		if !ok {
			return nil, fmt.Errorf("missing argument %s", arg)
		}

		argArr = append(argArr, val)
	}

	return d.engine.ExecuteProcedure(ctx, proc.Name, argArr,
		eng.WithCaller(opts.Caller),
		eng.WithDatasetID(d.DBID()),
	)
}

// Query executes a ad-hoc, read-only query.
func (d *Dataset) Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error) {
	return d.db.Query(ctx, stmt, args)
}

// ListProcedures returns the procedures in the dataset.
func (d *Dataset) ListProcedures() []*types.Procedure {
	var procs []*types.Procedure
	for _, procedure := range d.metadata.Procedures {
		procs = append(procs, procedure)
	}

	return procs
}

// ListTables returns the tables in the dataset.
func (d *Dataset) ListTables(ctx context.Context) ([]*types.Table, error) {
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
	var errs []error

	err := d.engine.Close()
	if err != nil {
		errs = append(errs, err)
	}

	err = d.db.Delete()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// Metadata returns the metadata for the dataset.
func (d *Dataset) Metadata() (name, owner string) {
	return d.options.name, d.options.owner
}
