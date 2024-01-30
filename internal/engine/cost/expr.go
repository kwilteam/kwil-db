package cost

import "fmt"

// LogicalExpr represents the strategies to access the required data.
// It's different from tree.Expression in that it will be used to access the data.
type LogicalExpr interface {
	fmt.Stringer

	// Resolve returns the field that this expression represents from the input
	// logical plan.
	Resolve(LogicalPlan) Field
}

// columnExpr represents a column in a schema.
// NOTE: it will be transformed to columnIdxExpr in the logical plan.????
type columnExpr struct {
	table string
	name  string
}

func (c *columnExpr) String() string {
	return c.name
}

func (c *columnExpr) Resolve(plan LogicalPlan) Field {
	for _, field := range plan.Schema().Fields {
		if field.Name == c.name {
			return field
		}
	}
	panic(fmt.Sprintf("field %s not found", c.name))
}

func Column(table, name string) LogicalExpr {
	return &columnExpr{table: table, name: name}
}

// columnIdxExpr represents a column in a schema by its index.
type columnIdxExpr struct {
	idx int
}

func (c *columnIdxExpr) String() string {
	return fmt.Sprintf("$%d", c.idx)
}

func (c *columnIdxExpr) Resolve(plan LogicalPlan) Field {
	return plan.Schema().Fields[c.idx]
}

func ColumnIdx(idx int) LogicalExpr {
	return &columnIdxExpr{idx: idx}
}

type aliasExpr struct {
	expr  LogicalExpr
	alias string
}

func (a *aliasExpr) String() string {
	return fmt.Sprintf("%s AS %s", a.expr, a.alias)
}

func (a *aliasExpr) Resolve(plan LogicalPlan) Field {
	return Field{Name: a.alias, Type: a.expr.Resolve(plan).Type}
}

func Alias(expr LogicalExpr, alias string) LogicalExpr {
	return &aliasExpr{expr: expr, alias: alias}
}

type literalStringExpr struct {
	value string
}

func (l *literalStringExpr) String() string {
	return l.value
}

func (l *literalStringExpr) Resolve(LogicalPlan) Field {
	return Field{Name: l.value, Type: "text"}
}

func LiteralString(value string) LogicalExpr {
	return &literalStringExpr{value: value}
}

type literalIntExpr struct {
	value int
}

func (l *literalIntExpr) String() string {
	return fmt.Sprintf("%d", l.value)
}

func (l *literalIntExpr) Resolve(LogicalPlan) Field {
	return Field{Name: fmt.Sprintf("%d", l.value), Type: "int"}
}

func LiteralInt(value int) LogicalExpr {
	return &literalIntExpr{value: value}
}

type OpExpr interface {
	LogicalExpr

	Op() string
}

type UnaryExpr interface {
	OpExpr
}

type unaryExpr struct {
	name string
	op   string
	expr LogicalExpr
}

func (n *unaryExpr) String() string {
	return fmt.Sprintf("%s %s", n.op, n.expr)
}

func (n *unaryExpr) Op() string {
	return n.op
}

func (n *unaryExpr) Resolve(LogicalPlan) Field {
	return Field{Name: n.name, Type: "bool"}
}

func Not(expr LogicalExpr) UnaryExpr {
	return &unaryExpr{
		name: "not",
		op:   "NOT",
		expr: expr,
	}
}

type BinaryExpr interface {
	OpExpr

	L() LogicalExpr
	R() LogicalExpr
}

type BoolBinaryExpr interface {
	BinaryExpr

	returnBool()
}

// boolBinaryExpr represents a binary expression that returns a boolean value.
type boolBinaryExpr struct {
	name string
	op   string
	l    LogicalExpr
	r    LogicalExpr
}

func (e *boolBinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.l, e.op, e.r)
}

func (e *boolBinaryExpr) Op() string {
	return e.op
}

func (e *boolBinaryExpr) L() LogicalExpr {
	return e.l
}

func (e *boolBinaryExpr) R() LogicalExpr {
	return e.r
}

func (e *boolBinaryExpr) Resolve(LogicalPlan) Field {
	return Field{Name: e.name, Type: "bool"}
}

func (e *boolBinaryExpr) returnBool() {}

func And(l, r LogicalExpr) BinaryExpr {
	return &boolBinaryExpr{
		name: "and",
		op:   "AND",
		l:    l,
		r:    r,
	}
}

func Or(l, r LogicalExpr) BinaryExpr {
	return &boolBinaryExpr{
		name: "or",
		op:   "OR",
		l:    l,
		r:    r,
	}
}

func Eq(l, r LogicalExpr) BinaryExpr {
	return &boolBinaryExpr{
		name: "eq",
		op:   "=",
		l:    l,
		r:    r,
	}
}

func Neq(l, r LogicalExpr) BinaryExpr {
	return &boolBinaryExpr{
		name: "neq",
		op:   "!=",
		l:    l,
		r:    r,
	}
}

func Gt(l, r LogicalExpr) BinaryExpr {
	return &boolBinaryExpr{
		name: "gt",
		op:   ">",
		l:    l,
		r:    r,
	}
}

func Gte(l, r LogicalExpr) BinaryExpr {
	return &boolBinaryExpr{
		name: "gte",
		op:   ">=",
		l:    l,
		r:    r,
	}
}

func Lt(l, r LogicalExpr) BinaryExpr {
	return &boolBinaryExpr{
		name: "lt",
		op:   "<",
		l:    l,
		r:    r,
	}
}

func Lte(l, r LogicalExpr) BinaryExpr {
	return &boolBinaryExpr{
		name: "lte",
		op:   "<=",
		l:    l,
		r:    r,
	}
}

type ArithmeticBinaryExpr interface {
	BinaryExpr

	returnOperandType()
}

// arithmeticBinaryExpr represents a binary expression that performs arithmetic
// operations, which return type of one of the operands.
type arithmeticBinaryExpr struct {
	name string
	op   string
	l    LogicalExpr
	r    LogicalExpr
}

func (e *arithmeticBinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.l, e.name, e.r)
}

func (e *arithmeticBinaryExpr) Op() string {
	return e.op
}

func (e *arithmeticBinaryExpr) L() LogicalExpr {
	return e.l
}

func (e *arithmeticBinaryExpr) R() LogicalExpr {
	return e.r
}

func (e *arithmeticBinaryExpr) Resolve(plan LogicalPlan) Field {
	return Field{Name: e.name, Type: e.l.Resolve(plan).Type}
}

func (e *arithmeticBinaryExpr) returnOperandType() {}

func Add(l, r LogicalExpr) BinaryExpr {
	return &arithmeticBinaryExpr{
		name: "add",
		op:   "+",
		l:    l,
		r:    r,
	}
}

func Sub(l, r LogicalExpr) BinaryExpr {
	return &arithmeticBinaryExpr{
		name: "sub",
		op:   "-",
		l:    l,
		r:    r,
	}
}

func Mul(l, r LogicalExpr) BinaryExpr {
	return &arithmeticBinaryExpr{
		name: "mul",
		op:   "*",
		l:    l,
		r:    r,
	}
}

func Div(l, r LogicalExpr) BinaryExpr {
	return &arithmeticBinaryExpr{
		name: "div",
		op:   "/",
		l:    l,
		r:    r,
	}
}

type AggregateExpr interface {
	LogicalExpr

	aggregate()
}

// aggregateExpr represents an aggregate expression.
// It returns a single value for a group of rows.
type aggregateExpr struct {
	name string
	expr LogicalExpr
	//NOTE add alias??
}

func (a *aggregateExpr) String() string {
	return fmt.Sprintf("%s(%s)", a.name, a.expr)
}

func (a *aggregateExpr) Resolve(plan LogicalPlan) Field {
	return Field{Name: a.name, Type: a.expr.Resolve(plan).Type}
}

func (a *aggregateExpr) aggregate() {}

func Max(expr LogicalExpr) AggregateExpr {
	return &aggregateExpr{name: "MAX", expr: expr}
}

func Min(expr LogicalExpr) AggregateExpr {
	return &aggregateExpr{name: "MIN", expr: expr}
}

func Avg(expr LogicalExpr) AggregateExpr {
	return &aggregateExpr{name: "AVG", expr: expr}
}

func Sum(expr LogicalExpr) AggregateExpr {
	return &aggregateExpr{name: "SUM", expr: expr}
}

// aggregateIntExpr represents an aggregate expression that returns an integer.
type aggregateIntExpr struct {
	name string
	expr LogicalExpr
}

func (a *aggregateIntExpr) String() string {
	return fmt.Sprintf("%s(%s)", a.name, a.expr)
}

func (a *aggregateIntExpr) Resolve(LogicalPlan) Field {
	return Field{Name: a.name, Type: "int"}
}

func (a *aggregateIntExpr) aggregate() {}

func Count(expr LogicalExpr) AggregateExpr {
	return &aggregateIntExpr{name: "COUNT", expr: expr}
}

type binaryExprBuilder interface {
	And(r LogicalExpr) BinaryExpr
	Or(r LogicalExpr) BinaryExpr
	Eq(r LogicalExpr) BinaryExpr
	Neq(r LogicalExpr) BinaryExpr
	Gt(r LogicalExpr) BinaryExpr
	Gte(r LogicalExpr) BinaryExpr
	Lt(r LogicalExpr) BinaryExpr
	Lte(r LogicalExpr) BinaryExpr

	Add(r LogicalExpr) BinaryExpr
	Sub(r LogicalExpr) BinaryExpr
	Mul(r LogicalExpr) BinaryExpr
	Div(r LogicalExpr) BinaryExpr
}

type binaryExprBuilderImpl struct {
	l LogicalExpr
}

func (b *binaryExprBuilderImpl) And(r LogicalExpr) BinaryExpr {
	return And(b.l, r)
}

func (b *binaryExprBuilderImpl) Or(r LogicalExpr) BinaryExpr {
	return Or(b.l, r)
}

func (b *binaryExprBuilderImpl) Eq(r LogicalExpr) BinaryExpr {
	return Eq(b.l, r)
}

func (b *binaryExprBuilderImpl) Neq(r LogicalExpr) BinaryExpr {
	return Neq(b.l, r)
}

func (b *binaryExprBuilderImpl) Gt(r LogicalExpr) BinaryExpr {
	return Gt(b.l, r)
}

func (b *binaryExprBuilderImpl) Gte(r LogicalExpr) BinaryExpr {
	return Gte(b.l, r)
}

func (b *binaryExprBuilderImpl) Lt(r LogicalExpr) BinaryExpr {
	return Lt(b.l, r)
}

func (b *binaryExprBuilderImpl) Lte(r LogicalExpr) BinaryExpr {
	return Lte(b.l, r)
}

func (b *binaryExprBuilderImpl) Add(r LogicalExpr) BinaryExpr {
	return Add(b.l, r)
}

func (b *binaryExprBuilderImpl) Sub(r LogicalExpr) BinaryExpr {
	return Sub(b.l, r)
}

func (b *binaryExprBuilderImpl) Mul(r LogicalExpr) BinaryExpr {
	return Mul(b.l, r)
}

func (b *binaryExprBuilderImpl) Div(r LogicalExpr) BinaryExpr {
	return Div(b.l, r)
}
