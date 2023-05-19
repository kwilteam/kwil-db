package engine2

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/engine2/utils"
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
	return e.db.CreateTable(ctx, datasetTable)
}

const (
	sqlListDatabases        = "SELECT name, owner FROM datasets;"
	sqlListDatabasesByOwner = "SELECT dbid, name, owner FROM datasets WHERE owner = $owner;"
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

func (e *engine) storeDataset(name, owner string) error {
	return e.db.Execute(sqlCreateDataset, map[string]any{
		"$dbid":  utils.GenerateDBID(name, owner),
		"$name":  name,
		"$owner": owner,
	})
}

func (e *engine) datasetExists(ctx context.Context, dbid string) (bool, error) {
	result, err := e.db.Query(ctx, sqlGetDataset, map[string]any{
		"$dbid": dbid,
	})
	if err != nil {
		return false, err
	}

	return len(result.Records()) > 0, nil
}

func (e *engine) deleteDataset(ctx context.Context, dbid string) error {
	return e.db.Execute(sqlDeleteDataset, map[string]any{
		"$dbid": dbid,
	})
}
