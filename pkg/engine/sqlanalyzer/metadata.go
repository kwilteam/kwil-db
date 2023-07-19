package sqlanalyzer

import "github.com/kwilteam/kwil-db/pkg/engine/types"

// schemaMetadata is used to store and retrieve metadata about the schema
type schemaMetadata struct {
	tables map[string]*types.Table
}

func newTableMetadata(tbl *types.Table) (*tableMetadata, error) {
	pks, err := tbl.GetPrimaryKey()
	if err != nil {
		return nil, err
	}

	return &tableMetadata{
		table:       tbl,
		primaryKeys: pks,
	}, nil
}

type tableMetadata struct {
	table       *types.Table
	primaryKeys []string
}
