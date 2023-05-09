package tree

type IndexedColumn struct {
	stmt       *indexedColumnBuilder
	Column     string
	Expression Expression
	Collation  CollationType
	OrderType  OrderType
}

func (i *IndexedColumn) ToSQL() string {
	i.check()

	i.stmt = Builder.BeginIndexedColumn()

	if i.Expression == nil {
		i.stmt.ColumnName(i.Column)
	} else {
		i.stmt.Expression(i.Expression)
	}

	if i.Collation != "" {
		i.stmt.CollationName(i.Collation)
	}

	if i.OrderType != OrderTypeNone {
		i.stmt.OrderType(i.OrderType)
	}

	return i.stmt.String()
}

func (i *IndexedColumn) check() {
	if i.Column == "" && i.Expression == nil {
		panic("column and expression cannot both be empty")
	}

	if i.Column != "" && i.Expression != nil {
		panic("column and expression cannot both be set")
	}
}

type indexedColumnBuilder struct {
	stmt *sqlBuilder
}

func (b *builder) BeginIndexedColumn() *indexedColumnBuilder {
	return &indexedColumnBuilder{
		stmt: newSQLBuilder(),
	}
}

func (b *indexedColumnBuilder) ColumnName(column string) {
	b.stmt.Write(SPACE)
	b.stmt.WriteIdent(column)
	b.stmt.Write(SPACE)
}

func (b *indexedColumnBuilder) Expression(expression Expression) {
	b.stmt.Write(SPACE, LPAREN, SPACE)
	b.stmt.WriteString(expression.ToSQL())
	b.stmt.Write(SPACE, RPAREN, SPACE)
}

func (b *indexedColumnBuilder) CollationName(collationName CollationType) {
	b.stmt.Write(SPACE, COLLATE, SPACE)
	b.stmt.WriteString(collationName.String())
	b.stmt.Write(SPACE)
}

func (b *indexedColumnBuilder) OrderType(orderType OrderType) {
	b.stmt.Write(SPACE)
	switch orderType {
	case OrderTypeAsc:
		b.stmt.Write(ASC)
	case OrderTypeDesc:
		b.stmt.Write(DESC)
	}
	b.stmt.Write(SPACE)
}

func (b *indexedColumnBuilder) String() string {
	return b.stmt.String()
}
