package dataset

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset/evaluater"
	"github.com/kwilteam/kwil-db/pkg/engine/execution"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

// A dataset is a deployed Kwil database with an underlying data store and engine.
type Dataset struct {
	metadata *Metadata
	db       Datastore
	engine   Engine
	log      log.Logger

	// initializers are the intiialization functions for extensions
	initializers map[string]Initializer
	// owner is the public key owner of the dataset
	owner User
	// name is the name of the dataset
	name string
	// allowMissingExtensions will let a dataset load, even if required extension initializers are not provided
	allowMissingExtensions bool
}

// OpenDataset opens a new dataset and loads the metadata from the database
func OpenDataset(ctx context.Context, ds Datastore, opts ...OpenOpt) (*Dataset, error) {
	dataset := &Dataset{
		db:                     ds,
		initializers:           make(map[string]Initializer),
		log:                    log.NewNoOp(),
		allowMissingExtensions: false,
	}

	for _, opt := range opts {
		opt(dataset)
	}

	engineOpts, err := dataset.getEngineOpts(ctx, dataset.initializers)
	if err != nil {
		return nil, err
	}

	eval, err := evaluater.NewEvaluater()
	if err != nil {
		return nil, err
	}

	engine, err := execution.NewEngine(ctx, datastoreWrapper{ds}, eval, engineOpts)
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
			_, err := d.executeOnce(ctx, procedure, []any{}, d.getExecutionOpts(procedure, opts)...)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}

const constructorName = "init"

// Execute executes a procedure.
func (d *Dataset) Execute(ctx context.Context, action string, args [][]any, opts *TxOpts) ([]map[string]any, error) {
	if opts == nil {
		opts = newTxOpts()
	}

	proc, err := d.getProcedure(action)
	if err != nil {
		return nil, err
	}

	if proc.RequiresAuthentication() && opts.Caller.PubKey() == nil {
		return nil, ErrCallerNotAuthenticated
	}

	if proc.IsOwnerOnly() && !bytes.Equal(opts.Caller.PubKey(), d.owner.PubKey()) {
		d.log.Debug("caller is not owner", zap.Binary("caller", opts.Caller.PubKey()), zap.Binary("owner", d.owner.PubKey()))
		return nil, ErrCallerNotOwner
	}

	savepoint, err := d.db.Savepoint()
	if err != nil {
		return nil, err
	}
	defer savepoint.Rollback()

	if len(args) == 0 { // if no args, add an empty arg map so we can execute once
		args = append(args, []any{})
	}

	var result []map[string]any
	for _, arg := range args {
		result, err = d.executeOnce(ctx, proc, arg, d.getExecutionOpts(proc, opts)...)
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

// Call is like execute, but it is non-mutative.
func (d *Dataset) Call(ctx context.Context, action string, args []any, opts *TxOpts) ([]map[string]any, error) {
	if opts == nil {
		opts = newTxOpts()
	}

	proc, err := d.getProcedure(action)
	if err != nil {
		return nil, err
	}
	if proc.IsMutative() {
		return nil, ErrCallMutativeProcedure
	}
	if proc.RequiresAuthentication() && opts.Caller.PubKey() == nil {
		return nil, ErrCallerNotAuthenticated
	}
	if proc.IsOwnerOnly() && !bytes.Equal(opts.Caller.PubKey(), d.owner.PubKey()) {
		return nil, ErrCallerNotOwner
	}

	if len(args) != len(proc.Args) {
		return nil, fmt.Errorf("expected %d args, got %d", len(proc.Args), len(args))
	}

	execOpts := append(d.getExecutionOpts(proc, opts), execution.CommittedOnly())
	return d.engine.ExecuteProcedure(ctx, proc.Name, args, execOpts...)
}

// getProcedure gets a procedure.  If the procedure is not found, it returns an error.
// if the procedure is a constructor/init procedure, it returns an error.
func (d *Dataset) getProcedure(action string) (*types.Procedure, error) {
	if strings.EqualFold(action, constructorName) {
		return nil, fmt.Errorf("cannot execute constructor")
	}

	proc, ok := d.metadata.Procedures[action]
	if !ok {
		return nil, fmt.Errorf("procedure %s does not exist", action)
	}

	return proc, nil
}

func (d *Dataset) getExecutionOpts(proc *types.Procedure, opts *TxOpts) []execution.ExecutionOpt {
	execOpts := []execution.ExecutionOpt{
		execution.WithDatasetID(d.DBID()),
	}
	if opts.Caller.PubKey() != nil {
		execOpts = append(execOpts, execution.WithCaller(opts.Caller))
	}

	if !proc.IsMutative() {
		execOpts = append(execOpts, execution.NonMutative())
	}

	return execOpts
}

func (d *Dataset) executeOnce(ctx context.Context, proc *types.Procedure, args []any, opts ...execution.ExecutionOpt) ([]map[string]any, error) {
	if len(args) != len(proc.Args) {
		return nil, fmt.Errorf("expected %d args, got %d", len(proc.Args), len(args))
	}

	sp, err := d.db.Savepoint()
	if err != nil {
		return nil, err
	}
	defer sp.Rollback()

	results, err := d.engine.ExecuteProcedure(ctx, proc.Name, args,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	err = sp.Commit()
	if err != nil {
		return nil, err
	}

	return results, nil
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

// ListExtensions returns the extensions in the dataset.
func (d *Dataset) ListExtensions(ctx context.Context) ([]*types.Extension, error) {
	return d.db.ListExtensions(ctx)
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
func (d *Dataset) Metadata() (name string, owner User) {
	return d.name, d.owner
}

func (d *Dataset) Savepoint() (sql.Savepoint, error) {
	return d.db.Savepoint()
}

func (d *Dataset) CreateSession() (sql.Session, error) {
	return d.db.CreateSession()
}

func (d *Dataset) ApplyChangeset(changeset io.Reader) error {
	return d.db.ApplyChangeset(changeset)
}
