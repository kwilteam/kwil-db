package planner2

import "github.com/kwilteam/kwil-db/core/types"

/*
	This file contains constants for cost, as well as logic for calculating the cost of things such
	as table scans, index searches, and chooses sorting algorithms.
*/

const (
	// ColumnAccessCost is the cost of accessing a column on disk.
	// It is NOT the cost of searching for a column.
	ColumnAccessCost = 1
)

// scanCost calculates the cost to scan a column
func scanCost(rowCount int64, rowTypes ...*types.DataType) int64 {
	panic("not implemented")
}

// indexSearchCost calculates the cost to search a column index
func indexSearchCost(rowCount int64, rowTypes ...*types.DataType) int64 {
	panic("not implemented")
}

// mergeJoinCost calculates the cost to merge join two columns
func mergeJoinCost(left *FilterCost, right *FilterCost) int64 {
	// if both are sargable, then we will build a hash table for the smaller one
}
