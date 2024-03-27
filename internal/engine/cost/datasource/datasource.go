package datasource

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type SourceType string

// DataSource represents a data source.
type DataSource interface {
	SchemaSource

	// SourceType returns the type of the data source.
	SourceType() SourceType

	// Scan scans the data source, return selected columns.
	// If projection field is not found in the schema, it will be ignored.
	// NOTE: should panic?
	Scan(ctx context.Context, projection ...string) *Result

	// Statistics returns the statistics of the data source.
	Statistics() *datatypes.Statistics
}

type DefaultSchemaSource struct {
	datasource DataSource
}

func (s *DefaultSchemaSource) Schema() *datatypes.Schema {
	return s.datasource.Schema()
}

func (s *DefaultSchemaSource) Scan(ctx context.Context, projection ...string) *Result {
	return s.datasource.Scan(ctx, projection...)
}

func DataAsSchemaSource(ds DataSource) SchemaSource {
	return &DefaultSchemaSource{datasource: ds}
}
