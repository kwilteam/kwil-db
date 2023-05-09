package pricing

import "github.com/kwilteam/kwil-db/pkg/databases/spec"

type PricingRequestType int

const (
	DEPLOY PricingRequestType = iota
	DROP
	QUERY
	WITHDRAW
)

type Params struct {
	Q spec.QueryType
	T int64 //Total number of Rows in the Table
	I int64 //Number of Indexed Columns in the Table
	S int64 //Size of each row in bytes
	U int64 //Number of rows that got updated due to the query operation
	W []int // List of  "where" predicate
}
