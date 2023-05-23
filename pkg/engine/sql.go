package engine

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
)

var datasetTable = &dto.Table{
	Name: "datasets",
	Columns: []*dto.Column{
		{
			Name: "dbid",
			Type: dto.TEXT,
			Attributes: []*dto.Attribute{
				{
					Type: dto.PRIMARY_KEY,
				},
				{
					Type: dto.NOT_NULL,
				},
			},
		},
		{
			Name: "owner",
			Type: dto.TEXT,
			Attributes: []*dto.Attribute{
				{
					Type: dto.NOT_NULL,
				},
			},
		},
		{
			Name: "name",
			Type: dto.TEXT,
			Attributes: []*dto.Attribute{
				{
					Type: dto.NOT_NULL,
				},
			},
		},
	},
	Indexes: []*dto.Index{
		{
			Name:    "datasets_owner_name",
			Columns: []string{"owner", "name"},
			Type:    dto.UNIQUE_BTREE,
		},
	},
}

func (e *engine) initTables(ctx context.Context) error {
	exists, err := e.db.TableExists(ctx, datasetTable.Name)
	if err != nil {
		return fmt.Errorf("failed to check if table %s exists: %w", datasetTable.Name, err)
	}

	if exists {
		return nil
	}

	return e.db.CreateTable(ctx, datasetTable)
}

const (
	sqlListDatabases        = "SELECT name, owner FROM datasets;"
	sqlListDatabasesByOwner = "SELECT name FROM datasets WHERE owner = $owner COLLATE NOCASE;"
	sqlDeleteDataset        = "DELETE FROM datasets WHERE dbid = $dbid;"
	sqlCreateDataset        = "INSERT INTO datasets (dbid, name, owner) VALUES ($dbid, $name, $owner);"
	sqlGetDataset           = "SELECT name, owner FROM datasets WHERE dbid = $dbid;"
)

func (e *engine) listDatasets(ctx context.Context) ([]struct {
	name  string
	owner string
}, error) {
	result, err := e.db.Query(ctx, sqlListDatabases, nil)
	if err != nil {
		return nil, err
	}

	var datasets []struct {
		name  string
		owner string
	}

	results := result.Records()
	for _, r := range results {
		datasets = append(datasets, struct {
			name  string
			owner string
		}{
			name:  r["name"].(string),
			owner: r["owner"].(string),
		})
	}

	return datasets, nil
}

// registerDataset registers a dataset in the master database.
func (e *engine) registerDataset(name, owner string) error {
	exists, err := e.datasetExists(context.Background(), utils.GenerateDBID(name, owner))
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("dataset %s already exists", name)
	}

	return e.db.Execute(sqlCreateDataset, map[string]any{
		"$dbid":  utils.GenerateDBID(name, owner),
		"$name":  name,
		"$owner": owner,
	})
}

// datasetExists checks if a dataset exists in the master database.
func (e *engine) datasetExists(ctx context.Context, dbid string) (bool, error) {
	result, err := e.db.Query(ctx, sqlGetDataset, map[string]any{
		"$dbid": dbid,
	})
	if err != nil {
		return false, err
	}

	return len(result.Records()) > 0, nil
}

// unregisterDataset unregisters a dataset from the master database.
func (e *engine) unregisterDataset(ctx context.Context, dbid string) error {
	return e.db.Execute(sqlDeleteDataset, map[string]any{
		"$dbid": dbid,
	})
}

// listDatasetsByOwner lists all datasets owned by a user.
func (e *engine) listDatasetsByOwner(ctx context.Context, owner string) ([]string, error) {
	result, err := e.db.Query(ctx, sqlListDatabasesByOwner, map[string]any{
		"$owner": owner,
	})
	if err != nil {
		return nil, err
	}

	var datasets []string

	results := result.Records()
	for _, r := range results {
		datasets = append(datasets, r["name"].(string))
	}

	return datasets, nil
}
