package master

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/utils"
)

type MasterDB struct {
	sqlStore *sqlStore
	path     string
	name     string
	DbidFunc DbidFunc
}

// New creates a new master database.
// It will initialize the master database if it does not exist.
func New(ctx context.Context, datastore Datastore, opts ...MasterOpt) (*MasterDB, error) {
	m := &MasterDB{
		sqlStore: &sqlStore{ds: datastore},
		path:     defaultPath,
		name:     defaultName,
		DbidFunc: utils.GenerateDBID,
	}

	for _, opt := range opts {
		opt(m)
	}

	err := m.sqlStore.init(ctx)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type DatasetInfo struct {
	DBID  string
	Name  string
	Owner string
}

// ListDatasets lists the dbids of all datasets.
func (d *MasterDB) ListDatasets(ctx context.Context) ([]*DatasetInfo, error) {
	return d.sqlStore.listDatasets(ctx)
}

// ListDatasetsByOwner lists all datasets owned by the given owner.
func (d *MasterDB) ListDatasetsByOwner(ctx context.Context, owner string) ([]string, error) {
	return d.sqlStore.listDatasetsByOwner(ctx, owner)
}

// RegisterDataset registers a new dataset.
func (d *MasterDB) RegisterDataset(ctx context.Context, name, owner string) error {
	dbid := d.DbidFunc(name, owner)

	exists, err := d.datasetExists(ctx, dbid)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("%w: %s", ErrDatasetExists, name)
	}

	return d.sqlStore.createDataset(ctx, dbid, name, owner)
}

// UnregisterDataset unregisters a dataset.
func (d *MasterDB) UnregisterDataset(ctx context.Context, dbid string) error {
	return d.sqlStore.deleteDataset(ctx, dbid)
}

func (d *MasterDB) datasetExists(ctx context.Context, dbid string) (bool, error) {
	ds, err := d.sqlStore.getDataset(ctx, dbid)
	if err != nil {
		return false, err
	}

	return ds != nil, nil
}

// Close closes the master database.
func (d *MasterDB) Close() error {
	return d.sqlStore.Close()
}
