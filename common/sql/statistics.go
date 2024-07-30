package sql

// NOTE: this file is TRANSITIONAL! These types are lifted from the
// unmerged internal/engine/costs/datatypes package.

import (
	"fmt"
	"strings"
)

// Statistics contains statistics about a table or a Plan. A Statistics can be
// derived directly from the underlying table, or derived from the statistics of
// its children.
type Statistics struct {
	RowCount int64

	ColumnStatistics []ColumnStatistics

	//Selectivity, for plan statistics
}

func (s *Statistics) String() string {
	var st strings.Builder
	fmt.Fprintf(&st, "RowCount: %d", s.RowCount)
	if len(s.ColumnStatistics) > 0 {
		fmt.Fprintln(&st, "")
	}
	for i, cs := range s.ColumnStatistics {
		fmt.Fprintf(&st, " Column %d:\n", i)
		fmt.Fprintf(&st, " - Min/Max = %v / %v\n", cs.Min, cs.Max)
		fmt.Fprintf(&st, " - NULL count = %v\n", cs.NullCount)
	}
	return st.String()
}

type ValCount struct {
	Val   any
	Count int
}

// ColumnStatistics contains statistics about a column.
type ColumnStatistics struct {
	NullCount int64

	Min      any
	MinCount int

	Max      any
	MaxCount int

	// MCVs are the most common values. It should be sorted by the value. It
	// should also be limited capacity, which means scan order has to be
	// deterministic since we have to throw out same-frequency observations.
	// (crap) Solution: multi-pass scan, merge lists, continue until no higher
	// freq values observed? OR when capacity reached, use a histogram? Do not
	// throw away MCVs, just start putting additional observations in to the
	// histogram instead.
	// MCVs []ValCount
	// MCVs map[cmp.Ordered]

	// MCVals  []any
	// MCFreqs []int

	// DistinctCount is harder. For example, unless we sub-sample
	// (deterministically), tracking distinct values could involve a data
	// structure with the same number of elements as rows in the table.
	DistinctCount int64

	AvgSize int64 // maybe: length of text, length of array, otherwise not used for scalar?

	// without histogram, we can make uniformity assumption to simplify the cost model
	//Histogram     []HistogramBucket
}
