package source

import "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"

// example.go contains code used in the documentation for the logical_plan package.

type exampleSchemaSource struct {
	schema *datatypes.Schema
}

func (s *exampleSchemaSource) Schema() *datatypes.Schema {
	return s.schema
}

func NewExampleSchemaSource(schema *datatypes.Schema) *exampleSchemaSource {
	return &exampleSchemaSource{schema: schema}
}
