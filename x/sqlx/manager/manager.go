package manager

import (
	"context"
	"fmt"
	"kwil/x/cfgx"
	"kwil/x/sqlx/cache"
	"kwil/x/sqlx/models"
	"kwil/x/sqlx/sqlclient"
)

/*
	The schema manager contains managers for interacting with the database.
	This includes a mix of abstractions, as well as a cache for the schema.
*/

type Manager struct {
	// The database client
	cache      Cache
	Client     *sqlclient.DB
	Metadata   MetadataManager
	Execution  ExecutionManager
	Deployment DeploymentManager
	Deposits   DepositsManager
}

type Cache interface {
	Store(db *models.Database) error
	Get(db string) *cache.Database // I know we shouldn't be coupling our code like this
}

// NewManager returns a new manager
func New(ctx context.Context, client *sqlclient.DB, cfg cfgx.Config, cache Cache) (*Manager, error) {
	mdm := NewMetadataManager(client)

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
	dbs, err := m.Metadata.GetAllDatabases(ctx)
	if err != nil {
		return err
	}

	for _, db := range dbs {
		bts, err := m.Metadata.GetMetadataBytes(ctx, db)
		if err != nil {
			continue
		}

		var modelDb models.Database
		err = modelDb.DecodeGOB(bts)
		if err != nil {
			continue
		}

		err = m.cache.Store(&modelDb)
		if err != nil {
			continue
		}
	}

	return nil
}

func (m *Manager) GetDatabase(ctx context.Context, db string) (*cache.Database, error) {
	cached := m.cache.Get(db)
	if cached == nil {
		return nil, fmt.Errorf("database %s not found", db)
	}

	return cached, nil

	// everything is held in memory currently, so we don't need to hit the database
	/*
		bts, err := m.Metadata.GetMetadataBytes(ctx, db)
		if err != nil {
			return nil,
		}

		var modelDb models.Database
		err = modelDb.DecodeGOB(bts)
		if err != nil {
			return nil, err
		}

		err = m.cache.Store(&modelDb)
		if err != nil {
			return nil, err
		}

		return m.cache.Get(db), nil
	*/
}
