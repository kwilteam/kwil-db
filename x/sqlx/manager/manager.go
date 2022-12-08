package manager

import (
	"context"
	"kwil/x/cfgx"
	"kwil/x/sqlx/sqlclient"
)

/*
	The schema manager contains managers for interacting with the database.
	This includes a mix of abstractions, as well as a cache for the schema.
*/

type Manager struct {
	// The database client
	cache      *SchemaCache
	Client     *sqlclient.DB
	Metadata   MetadataManager
	Execution  ExecutionManager
	Deployment DeploymentManager
	Deposits   DepositsManager
}

// NewManager returns a new manager
func New(ctx context.Context, client *sqlclient.DB, cfg cfgx.Config) (*Manager, error) {
	mdm := NewMetadataManager(client)

	// 60 minute interval
	cache := NewSchemaCache(mdm, 3600)
	go cache.RunGC(ctx)

	dpm, err := NewDepositsManager(client, cfg)
	if err != nil {
		return nil, err
	}

	return &Manager{
		cache:      cache,
		Client:     client,
		Metadata:   mdm,
		Deployment: NewDeploymentManager(cache, client),
		Execution:  NewExecutionManager(cache, client),
		Deposits:   dpm,
	}, nil
}

func (m *Manager) SyncCache(ctx context.Context) error {
	return m.cache.SyncAll(ctx)
}

// ExportDB returns the cachedDB for a given database
func (m *Manager) Get(dbName string) (*CachedDB, bool) {
	db, ok := m.cache.DBs[dbName]
	return db, ok
}

// Export returns the cachedDB for a given database
func (m *Manager) Export(dbName string) (*ExportedDB, error) {
	return m.cache.Export(dbName)
}
