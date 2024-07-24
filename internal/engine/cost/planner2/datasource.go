package planner2

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

type DataSource interface {
	// Schema returns the schema for the underlying data source
	Schema() *Schema
	// Statistics returns the statistics of the data source.
	Statistics() *datatypes.Statistics
}

type Catalog interface {
	GetDataSource(pgSchema, tableName string) (DataSource, error)
}

type EvalutationContext struct {
	Cost int64 // the running cost of the query
	// ???
}

type RelationStatistics struct {
	RowCount int64

	// ColumnStatistics is a map of column name to statistics.
	ColumnStatistics map[[2]string]*ColumnStatistics
	// ColumnOrder is a list of column names in the order they appear in the relation.
	// It also contains the parent relations name as the first element.
	ColumnOrder [][2]string
}

// Flatten aliases all columns with the given relation name.
// If conflicting column names are found, an error is returned.
func (r *RelationStatistics) Flatten(alias string) error {
	for i, col := range r.ColumnOrder {
		stats := r.ColumnStatistics[col]

		oldRelation := col[0]
		col[0] = alias

		_, ok := r.ColumnStatistics[col]
		if ok {
			return fmt.Errorf(`ambiguous column name: "%s"`, col[1])
		}

		r.ColumnOrder[i] = col
		r.ColumnStatistics[col] = stats

		delete(r.ColumnStatistics, [2]string{oldRelation, col[1]})
	}

	return nil
}

type ColumnStatistics struct {
	// basic stats
	NullCount int64
	Min       any
	Max       any

	// These are harder
	DistinctCount int64
	AvgSize       int64
}
