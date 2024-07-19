package datasource

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// memDataSource is a data source that reads data from memory.
type memDataSource struct {
	schema  *datatypes.Schema
	records []Row
}

func NewMemDataSource(s *datatypes.Schema, data []Row) *memDataSource {
	return &memDataSource{schema: s, records: data}
}

func (ds *memDataSource) Schema() *datatypes.Schema {
	return ds.schema
}

func (ds *memDataSource) Scan(ctx context.Context, projection ...string) *Result {
	return ScanData(ctx, ds.schema, ds.records, projection)
}

func (ds *memDataSource) Statistics() *datatypes.Statistics {
	panic("not implemented")
}

func (ds *memDataSource) SourceType() SourceType {
	return "memory"
}
