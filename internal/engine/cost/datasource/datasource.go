package datasource

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type SourceType string

// DataSource represents a data source.
type DataSource interface {
	// Schema returns the schema for the underlying data source
	Schema() *datatypes.Schema

	// SourceType returns the type of the data source.
	SourceType() SourceType

	// Scan scans the data source, return selected columns.
	// If projection field is not found in the schema, it will be ignored.
	// NOTE: should panic?
	Scan(ctx context.Context, projection ...string) *Result

	// Statistics returns the statistics of the data source.
	Statistics() *datatypes.Statistics
}
