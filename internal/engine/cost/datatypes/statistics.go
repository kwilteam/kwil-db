package datatypes

// Statistics contains statistics about a table.
type Statistics struct {
	RowCount int64

	ColumnStatistics []ColumnStatistic
}

// ColumnStatistic contains statistics about a column.
type ColumnStatistic struct {
	NullCount     int64
	Min           string
	Max           string
	DistinctCount int64
	AvgSize       int64

	// without histogram, we can make uniformity assumption to simplify the cost model
	//Histogram     []HistogramBucket
}

func (s *Statistics) ColStat(index int) *ColumnStatistic {
	return &s.ColumnStatistics[index]
}
