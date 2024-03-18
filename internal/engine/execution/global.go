package execution

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
)

// GlobalContext is the context for the entire execution.
// It exists for the lifetime of the server.
// It stores information about deployed datasets in-memory, and provides methods to interact with them.
type GlobalContext struct {
	// mu protects the datasets maps, which is written to during block execution
	// and read from during calls / queries.
	// It also implicitly protects maps held in the *baseDataset struct.
	mu sync.RWMutex

	// initializers are the namespaces that are available to datasets.
	// This includes other datasets, or loaded extensions.
	initializers map[string]precompiles.Initializer

	// datasets are the top level namespaces that are available to engine callers.
	// These only include datasets, and do not include extensions.
	datasets map[string]*baseDataset

	service *common.Service
}

var ErrDatasetNotFound = fmt.Errorf("dataset not found")

func InitializeEngine(ctx context.Context, tx sql.DB) error {
	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initTables,
	}

	err := versioning.Upgrade(ctx, tx, pg.InternalSchemaName, upgradeFns, engineVersion)
	if err != nil {
		return err
	}

	return nil
}

// NewGlobalContext creates a new global context. It will load any persisted
// datasets from the datastore. The provided database is only used for
// construction.
func NewGlobalContext(ctx context.Context, db sql.Executor, extensionInitializers map[string]precompiles.Initializer,
	service *common.Service) (*GlobalContext, error) {
	g := &GlobalContext{
		initializers: extensionInitializers,
		datasets:     make(map[string]*baseDataset),
		service:      service,
	}

	schemas, err := getSchemas(ctx, db)
	if err != nil {
		return nil, err
	}

	// we need to make sure schemas are ordered by their dependencies
	// if one schema is dependent on another, it must be loaded after the other
	// this is handled by the orderSchemas function
	for _, schema := range orderSchemas(schemas) {
		err := g.loadDataset(ctx, schema)
		if err != nil {
			return nil, err
		}
	}

	return g, nil
}

// CreateDataset deploys a schema.
// It will create the requisite tables, and perform the required initializations.
func (g *GlobalContext) CreateDataset(ctx context.Context, tx sql.DB, schema *common.Schema, caller []byte) (err error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	err = schema.Clean()
	if err != nil {
		return err
	}
	schema.Owner = caller

	err = g.loadDataset(ctx, schema)
	if err != nil {
		return err
	}

	err = createSchema(ctx, tx, schema)
	if err != nil {
		g.unloadDataset(schema.DBID())
		return err
	}

	return nil
}

// DeleteDataset deletes a dataset.
// It will ensure that the caller is the owner of the dataset.
func (g *GlobalContext) DeleteDataset(ctx context.Context, tx sql.DB, dbid string, caller []byte) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	dataset, ok := g.datasets[dbid]
	if !ok {
		return ErrDatasetNotFound
	}

	if !bytes.Equal(caller, dataset.schema.Owner) {
		return fmt.Errorf(`cannot delete dataset "%s", not owner`, dbid)
	}

	err := deleteSchema(ctx, tx, dbid)
	if err != nil {
		return err
	}

	g.unloadDataset(dbid)

	return nil
}

// Procedure calls a procedure on a dataset. It can be given either a readwrite or
// readonly transaction. If it is given a read-only transaction, it will not be
// able to execute any procedures that are not `view`.
func (g *GlobalContext) Procedure(ctx context.Context, tx sql.DB, options *common.ExecutionData) (*sql.ResultSet, error) {
	err := options.Clean()
	if err != nil {
		return nil, err
	}

	g.mu.RLock() // even if tx is readwrite, we will not change GlobalContext state, so we can use RLock
	defer g.mu.RUnlock()

	dataset, ok := g.datasets[options.Dataset]
	if !ok {
		return nil, ErrDatasetNotFound
	}

	procedureCtx := &precompiles.ProcedureContext{
		Ctx:       ctx,
		Signer:    options.Signer,
		Caller:    options.Caller,
		DBID:      options.Dataset,
		Procedure: options.Procedure,
		// starting with stack depth 0, increment in each action call
	}

	_, err = dataset.Call(procedureCtx, &common.App{
		Service: g.service,
		DB:      tx,
		Engine:  g,
	}, options.Procedure, options.Args)

	return procedureCtx.Result, err
}

// ListDatasets list datasets deployed by a specific caller.
// If caller is empty, it will list all datasets.
func (g *GlobalContext) ListDatasets(_ context.Context, caller []byte) ([]*types.DatasetIdentifier, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var datasets []*types.DatasetIdentifier
	if len(caller) == 0 { // prealloc only for all users' dataset
		datasets = make([]*types.DatasetIdentifier, 0, len(g.datasets))
	}
	for dbid, dataset := range g.datasets {
		if len(caller) == 0 || bytes.Equal(dataset.schema.Owner, caller) {
			datasets = append(datasets, &types.DatasetIdentifier{
				Name:  dataset.schema.Name,
				Owner: dataset.schema.Owner,
				DBID:  dbid,
			})
		}
	}

	return datasets, nil
}

// GetSchema gets a schema from a deployed dataset.
func (g *GlobalContext) GetSchema(_ context.Context, dbid string) (*common.Schema, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	dataset, ok := g.datasets[dbid]
	if !ok {
		return nil, ErrDatasetNotFound
	}

	return dataset.schema, nil
}

// Execute executes a SQL statement on a dataset. If the statement is mutative,
// the tx must also be a sql.AccessModer. It uses Kwil's SQL dialect.
func (g *GlobalContext) Execute(ctx context.Context, tx sql.DB, dbid, query string, values map[string]any) (*sql.ResultSet, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	dataset, ok := g.datasets[dbid]
	if !ok {
		return nil, ErrDatasetNotFound
	}

	// We have to parse the query and ensure the dbid is used to derive schema.
	// OR do we permit (or require) the schema to be specified in the query? It
	// could go either way, but this ad hoc query function is questionable anyway.
	parsed, err := sqlanalyzer.ApplyRules(query,
		sqlanalyzer.AllRules,
		dataset.schema.Tables, dbidSchema(dbid))
	if err != nil {
		return nil, err
	}

	if parsed.Mutative {
		txm, ok := tx.(sql.AccessModer)
		if !ok {
			return nil, errors.New("DB does not provide access mode needed for mutative statement")
		}
		if txm.AccessMode() == sql.ReadOnly {
			return nil, fmt.Errorf("cannot execute a mutative query in a read-only transaction")
		}
	}

	args := orderAndCleanValueMap(values, parsed.ParameterOrder)

	return tx.Execute(ctx, parsed.Statement, args...)
}

type dbQueryFn func(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)

// loadDataset loads a dataset into the global context.
// It does not create the dataset in the datastore.
func (g *GlobalContext) loadDataset(ctx context.Context, schema *common.Schema) error {
	dbid := schema.DBID()
	_, ok := g.initializers[dbid]
	if ok {
		return fmt.Errorf("dataset %s already exists", dbid)
	}

	datasetCtx := &baseDataset{
		schema:     schema,
		namespaces: make(map[string]precompiles.Instance),
		procedures: make(map[string]*procedure),
		global:     g,
	}

	for _, unprepared := range schema.Procedures {
		prepared, err := prepareProcedure(unprepared, g, schema)
		if err != nil {
			return err
		}

		_, ok := datasetCtx.procedures[prepared.name]
		if ok {
			return fmt.Errorf(`duplicate procedure name: "%s"`, prepared.name)
		}

		datasetCtx.procedures[prepared.name] = prepared
	}

	for _, ext := range schema.Extensions {
		_, ok := datasetCtx.namespaces[ext.Alias]
		if ok {
			return fmt.Errorf(`duplicate namespace assignment: "%s"`, ext.Alias)
		}

		initializer, ok := g.initializers[ext.Name]
		if !ok {
			return fmt.Errorf(`namespace "%s" not found`, ext.Name)
		}

		namespace, err := initializer(&precompiles.DeploymentContext{
			Ctx:    ctx,
			Schema: schema,
		}, g.service, ext.CleanMap())
		if err != nil {
			return err
		}

		datasetCtx.namespaces[ext.Alias] = namespace
	}

	g.initializers[dbid] = func(_ *precompiles.DeploymentContext, _ *common.Service, _ map[string]string) (precompiles.Instance, error) {
		return datasetCtx, nil
	}
	g.datasets[dbid] = datasetCtx

	return nil
}

// unloadDataset unloads a dataset from the global context.
// It does not delete the dataset from the datastore.
func (g *GlobalContext) unloadDataset(dbid string) {
	delete(g.datasets, dbid)
	delete(g.initializers, dbid)
}

// orderSchemas orders schemas based on their dependencies to other schemas.
func orderSchemas(schemas []*common.Schema) []*common.Schema {
	// Mapping from schema DBID to its extensions
	schemaMap := make(map[string][]string)
	for _, schema := range schemas {
		var exts []string
		for _, ext := range schema.Extensions {
			exts = append(exts, ext.Name)
		}
		schemaMap[schema.DBID()] = exts
	}

	// Topological sort
	var result []string
	visited := make(map[string]bool)
	var visitAll func(items []string)

	visitAll = func(items []string) {
		for _, item := range items {
			if !visited[item] {
				visited[item] = true
				visitAll(schemaMap[item])
				result = append(result, item)
			}
		}
	}

	keys := make([]string, 0, len(schemaMap))
	for key := range schemaMap {
		keys = append(keys, key)
	}
	sort.Strings(keys) // sort the keys for deterministic output
	visitAll(keys)

	// Reorder schemas based on result
	var orderedSchemas []*common.Schema
	for _, dbid := range result {
		for _, schema := range schemas {
			if schema.DBID() == dbid {
				orderedSchemas = append(orderedSchemas, schema)
				break
			}
		}
	}

	return orderedSchemas
}
