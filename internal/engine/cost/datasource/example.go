package datasource

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// example.go contains code used in the documentation for the logical_plan package.

type exampleDataSource struct {
	schema *datatypes.Schema
}

func (s *exampleDataSource) SourceType() SourceType {
	return "example"
}

func (s *exampleDataSource) Schema() *datatypes.Schema {
	return s.schema
}

func (s *exampleDataSource) Scan(ctx context.Context, projection ...string) *Result {
	panic("not implemented")
}

func (s *exampleDataSource) Statistics() *datatypes.Statistics {
	panic("not implemented")
}

func NewExampleDataSource(schema *datatypes.Schema) *exampleDataSource {
	return &exampleDataSource{schema: schema}
}
