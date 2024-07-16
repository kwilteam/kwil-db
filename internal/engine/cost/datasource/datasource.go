package datasource

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type SourceType string

type DataSource interface {
	// Schema returns the schema for the underlying data source
	Schema() *datatypes.Schema
	// Statistics returns the statistics of the data source.
	Statistics() *datatypes.Statistics
}

// DataSource represents a data source.
// NOTE: maybe should move to catalog package?
type FullDataSource interface {
	DataSource

	// SourceType returns the type of the data source.
	// SourceType() SourceType

	// Scan scans the data source, return selected columns. If projection field
	// is not found in the schema, it will be ignored. NOTE: should panic? This
	// method is like Execute, and it can't be implemented for an actual
	// postgres backend unless we really want to ask for `SCAN projection... `
	// with NO FILTERS.
	Scan(ctx context.Context, projection ...string) *Result
}
