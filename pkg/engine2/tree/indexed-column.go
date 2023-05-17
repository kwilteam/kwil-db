package tree

// * AFAIK this is a correct implementation of SQLITE's IndexedColumn, but it is not needed now

/*
type IndexedColumn struct {
	Column     string
	Expression Expression
	Collation  CollationType
	OrderType  OrderType
}

func (i *IndexedColumn) ToSQL() string {
	i.check()

	stmt := sqlwriter.NewWriter()

	if i.Expression == nil {
		stmt.WriteIdent(i.Column)
	} else {
		stmt.WriteString(i.Expression.ToSQL())
	}

	if i.Collation != "" {
		stmt.Token.Collate()
		stmt.WriteString(i.Collation.String())
	}

	if i.OrderType != OrderTypeNone {
		stmt.WriteString(i.OrderType.String())
	}

	return stmt.String()
}

func (i *IndexedColumn) check() {
	if i.Column == "" && i.Expression == nil {
		panic("column and expression cannot both be empty")
	}

	if i.Column != "" && i.Expression != nil {
		panic("column and expression cannot both be set")
	}
}
*/
