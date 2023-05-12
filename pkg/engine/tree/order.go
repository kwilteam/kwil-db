package tree

type OrderType string

const (
	OrderTypeNone OrderType = ""
	OrderTypeAsc  OrderType = "ASC"
	OrderTypeDesc OrderType = "DESC"
)

func (o OrderType) String() string {
	o.check()
	return string(o)
}

func (o OrderType) check() {
	if o != OrderTypeNone && o != OrderTypeAsc && o != OrderTypeDesc {
		panic("invalid order type")
	}
}
