package tree

type CTE struct {
	Table   string
	Columns []string
	Select  *Select
}

type OrderType string

const (
	OrderTypeNone OrderType = ""
	OrderTypeAsc  OrderType = "ASC"
	OrderTypeDesc OrderType = "DESC"
)

func (o OrderType) String() string {
	return string(o)
}
