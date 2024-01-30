package logical_plan

import (
	"fmt"
	dt "github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	pt "github.com/kwilteam/kwil-db/internal/engine/cost/plantree"
	tree "github.com/kwilteam/kwil-db/parse"
)

//
//type baseExprNode struct {
//	*pt.BaseTreeNode
//}
//
//func (e *baseExprNode) ExprNode() {}
//
//func newExprNode() *baseExprNode {
//	return &baseExprNode{BaseTreeNode: pt.NewBaseTreeNode()}
//}

// LogicalExpr represents the strategies to access the required data.
// It's different from tree.Expression in that it will be used to access the data.
type LogicalExpr interface {
	pt.TreeNode

	// Resolve returns the field that this expression represents from the schema
	Resolve(*dt.Schema) dt.Field
}

// ColumnExpr represents a column in a schema.
// NOTE: it will be transformed to columnIdxExpr in the logical plan.????
type ColumnExpr struct {
	*pt.BaseTreeNode

	Relation *dt.TableRef
	Name     string
}

var _ LogicalExpr = &ColumnExpr{}

func (e *ColumnExpr) String() string {
	if e.Relation != nil {
		return fmt.Sprintf("%s.%s", e.Relation.String(), e.Name)
	}
	return e.Name
}

func (e *ColumnExpr) Resolve(schema *dt.Schema) dt.Field {
	// TODO: use just one Column definition, right now we have:
	// - ColumnExpr
	// - dt.ColumnDef, to avoid circular import
	return schema.FieldFromColumn(dt.Column(e.Relation, e.Name))
}

// QualifyWithSchemas returns a new ColumnExpr with the relation set, i.e. qualified.
// NOTE:
// This feels like `Resolve`, but more coupled with implementation details.
// TODO: use all input's schemas as backup schemas
func (e *ColumnExpr) QualifyWithSchemas(schemas ...*dt.Schema) *ColumnExpr {
	if e.Relation != nil {
		return e
	}

	var schemaToUse *dt.Schema
	for _, schema := range schemas {
		var matchedFields []dt.Field
		for _, field := range schema.Fields {
			if field.Name == e.Name {
				matchedFields = append(matchedFields, field)
			}
		}

		switch len(matchedFields) {
		case 0:
			continue
		case 1:
			schemaToUse = schema
			break
		default:
			// handle ambiguous column, e.g. same column name in different tables
			// This can only happen when Join with USING clause, kwil doesn't support it yet.
			panic(fmt.Sprintf("cannot qualify ambiguous column: %s", e.Name))
		}
	}

	if schemaToUse == nil {
		panic(fmt.Sprintf("field %s not found", e.Name))
	}

	field := e.Resolve(schemaToUse)

	return &ColumnExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		Relation:     field.Relation(),
		Name:         field.Name,
	}
}

func ColumnUnqualified(name string) *ColumnExpr {
	return &ColumnExpr{BaseTreeNode: pt.NewBaseTreeNode(), Name: name}
}

func Column(table *dt.TableRef, name string) *ColumnExpr {
	return &ColumnExpr{BaseTreeNode: pt.NewBaseTreeNode(), Relation: table, Name: name}
}

// ColumnIdxExpr represents a column in a schema by its index.
type ColumnIdxExpr struct {
	*pt.BaseTreeNode

	Idx int
}

func (e *ColumnIdxExpr) String() string {
	return fmt.Sprintf("$%d", e.Idx)
}

func (e *ColumnIdxExpr) Resolve(schema *dt.Schema) dt.Field {
	return schema.Fields[e.Idx]
}

func ColumnIdx(idx int) LogicalExpr {
	return &ColumnIdxExpr{BaseTreeNode: pt.NewBaseTreeNode(), Idx: idx}
}

type AliasExpr struct {
	*pt.BaseTreeNode

	Relation *dt.TableRef
	Expr     LogicalExpr
	Alias    string
}

func (e *AliasExpr) String() string {
	return fmt.Sprintf("%s AS %s", e.Expr, e.Alias)
}

func (e *AliasExpr) Resolve(schema *dt.Schema) dt.Field {
	return dt.Field{Name: e.Alias, Type: e.Expr.Resolve(schema).Type}
}

func (e *AliasExpr) SetRelation(relation *dt.TableRef) {
	e.Relation = relation
}

func Alias(expr LogicalExpr, alias string) *AliasExpr {
	return &AliasExpr{BaseTreeNode: pt.NewBaseTreeNode(), Expr: expr, Alias: alias}
}

type LiteralExpr interface {
	LogicalExpr

	literal()
}

type LiteralTextExpr struct {
	*pt.BaseTreeNode

	Value string
}

func (e *LiteralTextExpr) String() string {
	return fmt.Sprintf("'%s'", e.Value)
}

func (e *LiteralTextExpr) Resolve(*dt.Schema) dt.Field {
	return dt.Field{Name: e.Value, Type: "text"}
}

func (e *LiteralTextExpr) literal() {}

func LiteralText(value string) *LiteralTextExpr {
	return &LiteralTextExpr{BaseTreeNode: pt.NewBaseTreeNode(), Value: value}
}

type LiteralNumericExpr struct {
	*pt.BaseTreeNode

	Value int64
}

func (e *LiteralNumericExpr) String() string {
	return fmt.Sprintf("%d", e.Value)
}

func (e *LiteralNumericExpr) Resolve(*dt.Schema) dt.Field {
	return dt.Field{Name: fmt.Sprintf("%d", e.Value), Type: "numeric"}
}

func (e *LiteralNumericExpr) literal() {}

func LiteralNumeric(value int64) *LiteralNumericExpr {
	return &LiteralNumericExpr{BaseTreeNode: pt.NewBaseTreeNode(), Value: value}
}

type LiteralBoolExpr struct {
	*pt.BaseTreeNode

	Value bool
}

func (e *LiteralBoolExpr) String() string {
	return fmt.Sprintf("%t", e.Value)
}

func (e *LiteralBoolExpr) Resolve(*dt.Schema) dt.Field {
	return dt.Field{Name: fmt.Sprintf("%t", e.Value), Type: "bool"}
}

func (e *LiteralBoolExpr) literal() {}

func LiteralBool(value bool) *LiteralBoolExpr {
	return &LiteralBoolExpr{BaseTreeNode: pt.NewBaseTreeNode(), Value: value}
}

type LiteralNullExpr struct {
	*pt.BaseTreeNode
}

func (e *LiteralNullExpr) String() string {
	return "NULL"
}

func (e *LiteralNullExpr) Resolve(*dt.Schema) dt.Field {
	return dt.Field{Name: "NULL", Type: "null"}
}

func (e *LiteralNullExpr) literal() {}

func LiteralNull() *LiteralNullExpr {
	return &LiteralNullExpr{BaseTreeNode: pt.NewBaseTreeNode()}
}

type LiteralBlobExpr struct {
	*pt.BaseTreeNode

	Value []byte
}

func (e *LiteralBlobExpr) String() string {
	return fmt.Sprintf("%x", e.Value)
}

func (e *LiteralBlobExpr) Resolve(*dt.Schema) dt.Field {
	return dt.Field{Name: fmt.Sprintf("%x", e.Value), Type: "blob"}
}

func (e *LiteralBlobExpr) literal() {}

func LiteralBlob(value []byte) *LiteralBlobExpr {
	return &LiteralBlobExpr{BaseTreeNode: pt.NewBaseTreeNode(), Value: value}
}

type OpExpr interface {
	LogicalExpr

	Op() string
}

type UnaryExpr interface {
	OpExpr

	E() LogicalExpr
}

type unaryExpr struct {
	*pt.BaseTreeNode

	name string
	op   string
	expr LogicalExpr
}

func (e *unaryExpr) String() string {
	return fmt.Sprintf("%s %s", e.op, e.expr)
}

func (e *unaryExpr) Op() string {
	return e.op
}

func (e *unaryExpr) Resolve(*dt.Schema) dt.Field {
	return dt.Field{Name: e.name, Type: "bool"}
}

func (e *unaryExpr) E() LogicalExpr {
	return e.expr
}

func Not(expr LogicalExpr) *unaryExpr {
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
	*pt.BaseTreeNode

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

func (e *boolBinaryExpr) Resolve(*dt.Schema) dt.Field {
	return dt.Field{Name: e.name, Type: "bool"}
}

func (e *boolBinaryExpr) returnBool() {}

func And(l, r LogicalExpr) *boolBinaryExpr {
	return &boolBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "and",
		op:           "AND",
		l:            l,
		r:            r,
	}
}

func Or(l, r LogicalExpr) *boolBinaryExpr {
	return &boolBinaryExpr{
		name: "or",
		op:   "OR",
		l:    l,
		r:    r,
	}
}

func Eq(l, r LogicalExpr) *boolBinaryExpr {
	return &boolBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "eq",
		op:           "=",
		l:            l,
		r:            r,
	}
}

func Neq(l, r LogicalExpr) *boolBinaryExpr {
	return &boolBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "neq",
		op:           "!=",
		l:            l,
		r:            r,
	}
}

func Gt(l, r LogicalExpr) *boolBinaryExpr {
	return &boolBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "gt",
		op:           ">",
		l:            l,
		r:            r,
	}
}

func Gte(l, r LogicalExpr) *boolBinaryExpr {
	return &boolBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "gte",
		op:           ">=",
		l:            l,
		r:            r,
	}
}

func Lt(l, r LogicalExpr) *boolBinaryExpr {
	return &boolBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "lt",
		op:           "<",
		l:            l,
		r:            r,
	}
}

func Lte(l, r LogicalExpr) *boolBinaryExpr {
	return &boolBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "lte",
		op:           "<=",
		l:            l,
		r:            r,
	}
}

type ArithmeticBinaryExpr interface {
	BinaryExpr

	returnOperandType()
}

// arithmeticBinaryExpr represents a binary expression that performs arithmetic
// operations, which return type of one of the operands.
type arithmeticBinaryExpr struct {
	*pt.BaseTreeNode

	name string
	op   string
	l    LogicalExpr
	r    LogicalExpr
}

func (e *arithmeticBinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.l, e.op, e.r)
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

func (e *arithmeticBinaryExpr) Resolve(schema *dt.Schema) dt.Field {
	return dt.Field{Name: e.name, Type: e.l.Resolve(schema).Type}
}

func (e *arithmeticBinaryExpr) returnOperandType() {}

func Add(l, r LogicalExpr) *arithmeticBinaryExpr {
	return &arithmeticBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "add",
		op:           "+",
		l:            l,
		r:            r,
	}
}

func Sub(l, r LogicalExpr) *arithmeticBinaryExpr {
	return &arithmeticBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "sub",
		op:           "-",
		l:            l,
		r:            r,
	}
}

func Mul(l, r LogicalExpr) *arithmeticBinaryExpr {
	return &arithmeticBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "mul",
		op:           "*",
		l:            l,
		r:            r,
	}
}

func Div(l, r LogicalExpr) *arithmeticBinaryExpr {
	return &arithmeticBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         "div",
		op:           "/",
		l:            l,
		r:            r,
	}
}

type AggregateExpr interface {
	LogicalExpr

	E() LogicalExpr
	aggregate()
}

// aggregateExpr represents an aggregate expression.
// It returns a single value for a group of rows.
type aggregateExpr struct {
	*pt.BaseTreeNode

	name string
	expr LogicalExpr
	//NOTE add alias??
}

func (e *aggregateExpr) String() string {
	return fmt.Sprintf("%s(%s)", e.name, e.expr)
}

func (e *aggregateExpr) Resolve(schema *dt.Schema) dt.Field {
	return dt.Field{Name: e.name, Type: e.expr.Resolve(schema).Type}
}

func (e *aggregateExpr) E() LogicalExpr {
	return e.expr
}

func (e *aggregateExpr) aggregate() {}

func Max(expr LogicalExpr) *aggregateExpr {
	return &aggregateExpr{BaseTreeNode: pt.NewBaseTreeNode(), name: "MAX", expr: expr}
}

func Min(expr LogicalExpr) *aggregateExpr {
	return &aggregateExpr{BaseTreeNode: pt.NewBaseTreeNode(), name: "MIN", expr: expr}
}

func Avg(expr LogicalExpr) *aggregateExpr {
	return &aggregateExpr{BaseTreeNode: pt.NewBaseTreeNode(), name: "AVG", expr: expr}
}

func Sum(expr LogicalExpr) *aggregateExpr {
	return &aggregateExpr{BaseTreeNode: pt.NewBaseTreeNode(), name: "SUM", expr: expr}
}

// aggregateIntExpr represents an aggregate expression that returns an integer.
type aggregateIntExpr struct {
	*pt.BaseTreeNode

	name string
	expr LogicalExpr
}

func (a *aggregateIntExpr) String() string {
	return fmt.Sprintf("%s(%s)", a.name, a.expr)
}

func (a *aggregateIntExpr) Resolve(*dt.Schema) dt.Field {
	return dt.Field{Name: a.name, Type: "int"}
}

func (a *aggregateIntExpr) E() LogicalExpr {
	return a.expr
}

func (a *aggregateIntExpr) aggregate() {}

func Count(expr LogicalExpr) *aggregateIntExpr {
	// TODO: Count should be an sql function expression
	return &aggregateIntExpr{BaseTreeNode: pt.NewBaseTreeNode(), name: "COUNT", expr: expr}
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

type SortExpression struct {
	*pt.BaseTreeNode

	Expr       LogicalExpr
	asc        bool
	nullsFirst bool
}

func (e *SortExpression) String() string {
	order := "DESC"
	if e.asc {
		order = "ASC"
	}
	nullsOrder := "NULLS LAST"
	if e.nullsFirst {
		nullsOrder = "NULLS FIRST"
	}
	return fmt.Sprintf("%s %s %s", e.Expr, order, nullsOrder)
}

func (e *SortExpression) Resolve(schema *dt.Schema) dt.Field {
	return e.Expr.Resolve(schema)
}

func SortExpr(expr LogicalExpr, asc, nullsFirst bool) *SortExpression {
	return &SortExpression{
		BaseTreeNode: pt.NewBaseTreeNode(),
		Expr:         expr,
		asc:          asc,
		nullsFirst:   nullsFirst,
	}
}

type FuncExpr interface {
	LogicalExpr

	FuncName() string
}

type ScalarFuncExpr interface {
	FuncExpr

	scalarFn()
}

type AggregateFuncExpr interface {
	FuncExpr

	aggregateFn()
}

type scalarFuncExpr struct {
	*pt.BaseTreeNode

	fn   *tree.ExpressionFunctionCall // could be built-in or user-defined
	args []LogicalExpr
}

func (e *scalarFuncExpr) String() string {
	return fmt.Sprintf("%s(%s)", e.fn.FunctionName(), PpList(e.args))
}

func (e *scalarFuncExpr) Resolve(schema *dt.Schema) dt.Field {
	panic("implement me")
}

func (e *scalarFuncExpr) FuncName() string {
	return e.fn.FunctionName()
}

func (e *scalarFuncExpr) scalarFn() {}

func ScalarFunc(fn *tree.ExpressionFunctionCall, args ...LogicalExpr) *scalarFuncExpr {
	return &scalarFuncExpr{BaseTreeNode: pt.NewBaseTreeNode(), fn: fn, args: args}
}

type aggregateFuncExpr struct {
	*pt.BaseTreeNode

	fn       *tree.ExpressionFunctionCall // could be built-in or user-defined
	args     []LogicalExpr
	distinct bool

	// TODO?
	filter  LogicalExpr   // optional: having
	orderBy []LogicalExpr // optional: optimization
}

func (e *aggregateFuncExpr) String() string {
	return fmt.Sprintf("%s(%s)", e.fn.FunctionName(), PpList(e.args))
}

func (e *aggregateFuncExpr) Resolve(schema *dt.Schema) dt.Field {
	panic("implement me")
}

func (e *aggregateFuncExpr) FuncName() string {
	return e.fn.FunctionName()
}

func (e *aggregateFuncExpr) aggregateFn() {}

func AggregateFunc(fn *tree.ExpressionFunctionCall, args []LogicalExpr, distinct bool,
	filter LogicalExpr, orderBy ...LogicalExpr) *aggregateFuncExpr {
	return &aggregateFuncExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		fn:           fn,
		args:         args,
		distinct:     distinct,
		filter:       filter,
		orderBy:      orderBy,
	}
}

//type CaseExpr struct

//// pt.TreeNode implementation
// Children() implementation

func (e *ColumnExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{}
}

func (e *ColumnIdxExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{}
}

func (e *AliasExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{e.Expr}
}

func (e *LiteralTextExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{}
}

func (e *LiteralNumericExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{}
}

func (e *unaryExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{e.expr}
}

func (e *boolBinaryExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{e.l, e.r}
}

func (e *arithmeticBinaryExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{e.l, e.r}
}

func (e *aggregateExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{e.expr}
}

func (e *aggregateIntExpr) Children() []pt.TreeNode {
	return []pt.TreeNode{e.expr}
}

func (e *SortExpression) Children() []pt.TreeNode {
	return []pt.TreeNode{e.Expr}
}

func (e *scalarFuncExpr) Children() []pt.TreeNode {
	return exprListToNodeList(e.args)
}

func (e *aggregateFuncExpr) Children() []pt.TreeNode {
	return exprListToNodeList(e.args)
}

// TransformChildren() implementation

func (e *ColumnExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return e
}

func (e *ColumnIdxExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return e
}

func (e *AliasExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &AliasExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		Expr:         fn(e.Expr).(LogicalExpr),
		Alias:        e.Alias,
	}
}

func (e *LiteralTextExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return e
}

func (e *LiteralNumericExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return e
}

func (e *unaryExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &unaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         e.name,
		op:           e.op,
		expr:         fn(e.expr).(LogicalExpr),
	}
}

func (e *boolBinaryExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &boolBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         e.name,
		op:           e.op,
		l:            fn(e.l).(LogicalExpr),
		r:            fn(e.r).(LogicalExpr),
	}
}

func (e *arithmeticBinaryExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &arithmeticBinaryExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         e.name,
		op:           e.op,
		l:            fn(e.l).(LogicalExpr),
		r:            fn(e.r).(LogicalExpr),
	}
}

func (e *aggregateExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &aggregateExpr{

		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         e.name,
		expr:         fn(e.expr).(LogicalExpr),
	}
}

func (e *aggregateIntExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &aggregateIntExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		name:         e.name,
		expr:         fn(e.expr).(LogicalExpr),
	}
}

func (e *SortExpression) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	return &SortExpression{
		BaseTreeNode: pt.NewBaseTreeNode(),
		Expr:         fn(e.Expr).(LogicalExpr),
		asc:          e.asc,
		nullsFirst:   e.nullsFirst,
	}
}

func (e *scalarFuncExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	var newArgs []LogicalExpr
	for _, arg := range e.args {
		newArgs = append(newArgs, fn(arg).(LogicalExpr))
	}

	return &scalarFuncExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		fn:           e.fn,
		args:         newArgs,
	}
}

func (e *aggregateFuncExpr) TransformChildren(fn pt.TransformFunc) pt.TreeNode {
	var newArgs []LogicalExpr
	for _, arg := range e.args {
		newArgs = append(newArgs, fn(arg).(LogicalExpr))
	}

	return &aggregateFuncExpr{
		BaseTreeNode: pt.NewBaseTreeNode(),
		fn:           e.fn,
		args:         newArgs,
		distinct:     e.distinct,
		//filter:       fn(e.filter).(LogicalExpr),
		//orderBy:      ,
	}
}

// ExprNode() implementation

func (e *ColumnExpr) ExprNode()           {}
func (e *ColumnIdxExpr) ExprNode()        {}
func (e *AliasExpr) ExprNode()            {}
func (e *LiteralTextExpr) ExprNode()      {}
func (e *LiteralNumericExpr) ExprNode()   {}
func (e *unaryExpr) ExprNode()            {}
func (e *boolBinaryExpr) ExprNode()       {}
func (e *arithmeticBinaryExpr) ExprNode() {}
func (e *aggregateExpr) ExprNode()        {}
func (e *aggregateIntExpr) ExprNode()     {}
func (e *SortExpression) ExprNode()       {}
