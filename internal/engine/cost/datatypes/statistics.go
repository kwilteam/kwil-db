package datatypes

import "fmt"

// Statistics contains statistics about a table or a Plan. A Statistics can be
// derived directly from the underlying table, or derived from the statistics of
// its children.
type Statistics struct {
	RowCount int64

	ColumnStatistics []ColumnStatistics

	//Selectivity, for plan statistics
}

func (s *Statistics) String() string {
	return fmt.Sprintf("RowCount: %d", s.RowCount)
}

// ColumnStatistics contains statistics about a column.
type ColumnStatistics struct {
	NullCount     int64
	Min           any
	Max           any
	DistinctCount int64
	AvgSize       int64

	// without histogram, we can make uniformity assumption to simplify the cost model
	//Histogram     []HistogramBucket
}

func (s *Statistics) ColStat(index int) *ColumnStatistics {
	return &s.ColumnStatistics[index]
}

func NewStatistics(rowCount int64, colStats []ColumnStatistics) *Statistics {
	return &Statistics{
		RowCount:         rowCount,
		ColumnStatistics: colStats,
	}
}

func NewEmptyStatistics() *Statistics {
	return &Statistics{
		RowCount:         0,
		ColumnStatistics: nil,
	}
}
