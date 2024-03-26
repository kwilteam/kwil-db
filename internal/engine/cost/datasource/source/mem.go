package source

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// memDataSource is a data source that reads data from memory.
type memDataSource struct {
	schema  *datatypes.Schema
	records []datasource.Row
}

func NewMemDataSource(s *datatypes.Schema, data []datasource.Row) *memDataSource {
	return &memDataSource{schema: s, records: data}
}

func (ds *memDataSource) Schema() *datatypes.Schema {
	return ds.schema
}

func (ds *memDataSource) Scan(projection ...string) *datasource.Result {
	return datasource.dsScan(ds.schema, ds.records, projection)
}

func (ds *memDataSource) SourceType() datasource.SourceType {
	return "memory"
}
