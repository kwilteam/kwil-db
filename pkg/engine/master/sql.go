package master

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

type sqlStore struct {
	ds Datastore
}

const (
	datasetsTableName       = "datasets"
	sqlCreateDatasetTable   = "CREATE TABLE IF NOT EXISTS " + datasetsTableName + " (dbid TEXT PRIMARY KEY NOT NULL, name TEXT NOT NULL, owner BLOB NOT NULL) WITHOUT ROWID, STRICT;"
	sqlCreateDatasetIndex   = "CREATE UNIQUE INDEX IF NOT EXISTS idx_datasets_owner_name ON " + datasetsTableName + " (owner, name);"
	sqlListDatasets         = "SELECT dbid, name, owner FROM " + datasetsTableName + ";"
	sqlListDatabasesByOwner = "SELECT name FROM " + datasetsTableName + " WHERE public_key(owner) = $owner;"
	sqlDeleteDataset        = "DELETE FROM " + datasetsTableName + " WHERE dbid = $dbid;"
	sqlCreateDataset        = "INSERT INTO " + datasetsTableName + " (dbid, name, owner) VALUES ($dbid, $name, $owner);"
	sqlGetDataset           = "SELECT name, owner FROM " + datasetsTableName + " WHERE dbid = $dbid;"
)

// init will create the "datasets" table if it does not exist
func (d *sqlStore) init(ctx context.Context) error {
	exists, err := d.ds.TableExists(ctx, datasetsTableName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	sp, err := d.ds.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	// create dataset table
	err = d.ds.Execute(ctx, sqlCreateDatasetTable, nil)
	if err != nil {
		return err
	}

	// create dataset index
	err = d.ds.Execute(ctx, sqlCreateDatasetIndex, nil)
	if err != nil {
		return err
	}

	return sp.Commit()
}

func (d *sqlStore) getDataset(ctx context.Context, dbid string) (*types.DatasetInfo, error) {
	results, err := d.ds.Query(ctx, sqlGetDataset, map[string]any{
		"$dbid": dbid,
	})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}
	name, ok := results[0]["name"].(string)
	if !ok {
		return nil, fmt.Errorf("error getting dataset name from result set")
	}

	ownerBts, ok := results[0]["owner"].([]byte)
	if !ok {
		return nil, fmt.Errorf("error getting dataset owner fromr result set")
	}

	owner := &types.User{}
	err = owner.UnmarshalBinary(ownerBts)
	if err != nil {
		return nil, err
	}

	return &types.DatasetInfo{
		DBID:  dbid,
		Name:  name,
		Owner: owner,
	}, nil
}

func (d *sqlStore) listDatasetsByOwner(ctx context.Context, ownerPubKey []byte) ([]string, error) {
	results, err := d.ds.Query(ctx, sqlListDatabasesByOwner, map[string]any{
		"$owner": ownerPubKey,
	})
	if err != nil {
		return nil, err
	}

	var names []string
	for _, result := range results {
		name, ok := result["name"].(string)
		if !ok {
			return nil, fmt.Errorf("error getting dataset name from result set")
		}

		names = append(names, name)
	}

	return names, nil
}

func (d *sqlStore) createDataset(ctx context.Context, dbid, name string, owner *types.User) error {
	ownerBts, err := owner.MarshalBinary()
	if err != nil {
		return err
	}

	return d.ds.Execute(ctx, sqlCreateDataset, map[string]any{
		"$dbid":  dbid,
		"$name":  name,
		"$owner": ownerBts,
	})
}

func (d *sqlStore) deleteDataset(ctx context.Context, dbid string) error {
	return d.ds.Execute(ctx, sqlDeleteDataset, map[string]any{
		"$dbid": dbid,
	})
}

func (s *sqlStore) listDatasets(ctx context.Context) ([]*types.DatasetInfo, error) {
	results, err := s.ds.Query(ctx, sqlListDatasets, nil)
	if err != nil {
		return nil, err
	}

	var data []*types.DatasetInfo
	for _, result := range results {
		dbid, ok := result["dbid"].(string)
		if !ok {
			return nil, fmt.Errorf("error getting dbid from result set")
		}

		ownerBts, ok := result["owner"].([]byte)
		if !ok {
			return nil, fmt.Errorf("error getting owner from result set")
		}

		owner := &types.User{}
		err = owner.UnmarshalBinary(ownerBts)
		if err != nil {
			return nil, err
		}

		name, ok := result["name"].(string)
		if !ok {
			return nil, fmt.Errorf("error getting name from result set")
		}

		data = append(data, &types.DatasetInfo{
			DBID:  dbid,
			Name:  name,
			Owner: owner,
		})
	}

	return data, nil
}

func (d *sqlStore) Close() error {
	return d.ds.Close()
}
