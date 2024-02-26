package execution

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	coreTypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	sql "github.com/kwilteam/kwil-db/internal/sql"
)

// GlobalContext is the context for the entire execution.
// It exists for the lifetime of the server.
// It stores information about deployed datasets in-memory, and provides methods to interact with them.
type GlobalContext struct {
	// initializers are the namespaces that are available to datasets.
	// This includes other datasets, or loaded extensions.
	initializers map[string]ExtensionInitializer

	// datasets are the top level namespaces that are available to engine callers.
	// These only include datasets, and do not include extensions.
	datasets map[string]*baseDataset
}

// NewGlobalContext creates a new global context.
// It will load any persisted datasets from the datastore.
func NewGlobalContext(ctx context.Context, tx sql.DB, extensionInitializers map[string]ExtensionInitializer) (*GlobalContext, error) {
	g := &GlobalContext{
		initializers: extensionInitializers,
		datasets:     make(map[string]*baseDataset),
	}

	err := createSchemasTableIfNotExists(ctx, tx)
	if err != nil {
		return nil, err
	}

	schemas, err := getSchemas(ctx, tx)
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
func (g *GlobalContext) CreateDataset(ctx context.Context, tx sql.DB, schema *types.Schema, caller []byte) (err error) {
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
	dataset, ok := g.datasets[dbid]
	if !ok {
		return types.ErrDatasetNotFound
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

// Execute executes a procedure.
// It has the ability to mutate state, including uncommitted state. <=== UNCOMMITTED STATE, but caller is in various contexts (e.g. tx exec vs. RPC call)
// once we fix auth, signer should get removed, as they would be the same.
func (g *GlobalContext) Execute(ctx context.Context, tx sql.DB, options *types.ExecutionData) (*sql.ResultSet, error) {
	err := options.Clean()
	if err != nil {
		return nil, err
	}

	dataset, ok := g.datasets[options.Dataset]
	if !ok {
		return nil, types.ErrDatasetNotFound
	}

	procedureCtx := &ProcedureContext{
		Ctx:       ctx,
		Signer:    options.Signer,
		Caller:    options.Caller,
		globalCtx: g,
		values:    make(map[string]any),
		DBID:      options.Dataset,
		Procedure: options.Procedure,
		DB:        tx,
	}

	_, err = dataset.Call(procedureCtx, options.Procedure, options.Args)

	return procedureCtx.Result, err
}

// ListDatasets list datasets deployed by a specific caller.
// If caller is empty, it will list all datasets.
func (g *GlobalContext) ListDatasets(ctx context.Context, caller []byte) ([]*coreTypes.DatasetIdentifier, error) {
	datasets := make([]*coreTypes.DatasetIdentifier, 0, len(g.datasets))
	for dbid, dataset := range g.datasets {
		if len(caller) == 0 || bytes.Equal(dataset.schema.Owner, caller) {
			datasets = append(datasets, &coreTypes.DatasetIdentifier{
				Name:  dataset.schema.Name,
				Owner: dataset.schema.Owner,
				DBID:  dbid,
			})
		}
	}

	return datasets, nil
}

// GetSchema gets a schema from a deployed dataset.
func (g *GlobalContext) GetSchema(ctx context.Context, dbid string) (*types.Schema, error) {
	dataset, ok := g.datasets[dbid]
	if !ok {
		return nil, types.ErrDatasetNotFound
	}

	return dataset.schema, nil
}

// Query executes a read-only query.
func (g *GlobalContext) Query(ctx context.Context, tx sql.DB, dbid string, query string) (*sql.ResultSet, error) {
	dataset, ok := g.datasets[dbid] // data race with txsvc hitting this freely?
	if !ok {
		return nil, types.ErrDatasetNotFound
	}

	// We have to parse the query and ensure the dbid is used to derive schema.
	// OR do we permit (or require) the schema to be specified in the query? It
	// could go either way, but this ad hoc query function is questionable anyway.
	parsed, err := sqlanalyzer.ApplyRules(query,
		sqlanalyzer.NoCartesianProduct,
		dataset.schema.Tables, dbidSchema(dbid))
	if err != nil {
		return nil, err
	}

	if parsed.Mutative {
		return nil, fmt.Errorf("query is mutative")
	}

	return tx.Execute(ctx, parsed.Statement)
}

type dbQueryFn func(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)

// loadDataset loads a dataset into the global context.
// It does not create the dataset in the datastore.
func (g *GlobalContext) loadDataset(ctx context.Context, schema *types.Schema) error {
	dbid := schema.DBID()
	_, ok := g.initializers[dbid]
	if ok {
		return fmt.Errorf("dataset %s already exists", dbid)
	}

	datasetCtx := &baseDataset{
		schema:     schema,
		namespaces: make(map[string]ExtensionNamespace),
		procedures: make(map[string]*procedure),
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

		namespace, err := initializer(&DeploymentContext{
			Ctx:    ctx,
			Schema: schema,
		}, ext.CleanMap())
		if err != nil {
			return err
		}

		datasetCtx.namespaces[ext.Alias] = namespace
	}

	g.initializers[dbid] = func(_ *DeploymentContext, _ map[string]string) (ExtensionNamespace, error) {
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
func orderSchemas(schemas []*types.Schema) []*types.Schema {
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
	var orderedSchemas []*types.Schema
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
