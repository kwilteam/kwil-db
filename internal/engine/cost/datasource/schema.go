package datasource

import (
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// SchemaSource is an interface that provides the access to schema.
// It's used to get the schema of a table. It doesn't have the ability to
// scan the underlying data, which DataSource has.
type SchemaSource interface {
	// Schema returns the schema for the underlying data source
	Schema() *datatypes.Schema
}

// SchemaSourceToDataSource converts a SchemaSource to a DataSource.
func SchemaSourceToDataSource(ss SchemaSource) DataSource {
	switch t := ss.(type) {
	case *DefaultSchemaSource:
		return t.datasource
	case *csvDataSource:
		return t
	default:
		panic("SchemaSourceToDataSource: SchemaSource cannot be converted to DataSource")
	}
}
