package demo

import (
	"fmt"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"strings"
)

type LogicalScan struct {
	ds DataSource
	//projection []string
	schema *schema
}

func NewLogicalScan(ds DataSource) *LogicalScan {
	p := &LogicalScan{
		ds: ds,
		//projection: projection,
	}

	//p.schema = p.deriveSchema()
	p.schema = ds.Schema()
	return p
}

func (p *LogicalScan) Schema() *schema {
	return p.schema
}

func (p *LogicalScan) String() string {
	//if p.projection == nil {
	//	return "Scan: projection=*"
	//} else {
	//	return "Scan: projection=" + strings.Join(p.projection, ",")
	//}

	return "Scan: table=" + p.schema.tblName
}

func (p *LogicalScan) Inputs() []LogicalPlan {
	return nil
}

func (p *LogicalScan) Accept(v LogicalVisitor) any {
	return v.VisitLogicalScan(p)
}

func (p *LogicalScan) deriveSchema() *schema {
	//return p.ds.Schema().Select(p.projection)
	return p.ds.Schema().Select(nil)
}

type LogicalProjection struct {
	input LogicalPlan
	exprs []tree.ResultColumnExpression
}

func NewLogicalProjection(input LogicalPlan, exprs []tree.ResultColumnExpression) *LogicalProjection {
	return &LogicalProjection{
		input: input,
		exprs: exprs,
	}
}

func exprToField(expr tree.Expression, input LogicalPlan) *field {
	switch t := expr.(type) {
	case *tree.ExpressionColumn:
		for _, field := range input.Schema().fields {
			if field.ColName == t.Column {
				return field
			}
		}
	case *tree.ExpressionLiteral:
		return &field{
			ColName: t.Value,
			//Type:    t.Type,
			Type: "text",
		}
	case *tree.ExpressionFunction:
		var retType string
		switch t.Function.Name() {
		case "count", "min", "max":
			retType = "int"
		case "concat", "substr",
		}
		return &field{
			ColName:         t.Function.Name(),
			OriginalColName: t.Function.Name(),
			//Type: "int",
		}
	case *tree.ExpressionCase:
		return &field{
			ColName: "case",
		}
	case *tree.ExpressionArithmetic:
		return &field{
			ColName: t.Operator.String(),
			Type
		}

	default:
		panic("not implemented")
	}

	return nil
}

func (p *LogicalProjection) Schema() *schema {
	//return Schema(expr.map { it.toField(input) })

	// projection should return a new schema based on the input schema
	//return p.input.Schema().Select(p.exprs)
	cols := make([]*field, len(p.exprs))
	for _, expr := range p.exprs {
		cols = append(cols, exprToField(expr.Expression, p.input))
	}
	return newSchema(cols...)
}

func (p *LogicalProjection) String() string {
	fields := make([]string, len(p.exprs))
	for i, expr := range p.exprs {
		fields[i] = expr.ToSQL()
	}
	return "Projection: " + strings.Join(fields, ",")
	//return "Projection: " + strings.Join(p.exprs, ",")
}

func (p *LogicalProjection) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalProjection) Accept(v LogicalVisitor) any {
	return v.VisitLogicalProjection(p)
}

type LogicalFilter struct {
	input LogicalPlan
	expr  tree.Expression
}

func NewLogicalFilter(input LogicalPlan, expr tree.Expression) *LogicalFilter {
	return &LogicalFilter{
		input: input,
		expr:  expr,
	}
}

func (p *LogicalFilter) Schema() *schema {
	// filter won't change schema
	return p.input.Schema()
}

func (p *LogicalFilter) String() string {
	return "Filter: " + p.expr.ToSQL()
}

func (p *LogicalFilter) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalFilter) Accept(v LogicalVisitor) any {
	return v.VisitLogicalFilter(p)
}

type LogicalAggregate struct {
	input LogicalPlan
	// select a,b,agg(c) from t group by a,b;
	// the output schema always contain all group by columns
	// groupBy should be 'a,b'
	groupBy []tree.Expression
	//aggregates []tree.Expression
	// aggregates should be 'agg(c)'
	aggregates []*tree.AggregateFunc
}

func NewLogicalAggregate(input LogicalPlan, groupBy []tree.Expression,
	aggregates []*tree.AggregateFunc) *LogicalAggregate {
	return &LogicalAggregate{
		input:      input,
		groupBy:    groupBy,
		aggregates: aggregates,
	}
}

func (p *LogicalAggregate) Schema() *schema {
	// aggregate should return a new schema based on the input schema
	//return p.input.Schema().Select(p.groupBy)
	panic("not implemented")

	cols := make([]*field, len(p.groupBy)+len(p.aggregates))
	for i, expr := range p.groupBy {
		cols[i] = exprToField(expr, p.input)
	}

	for i, expr := range p.aggregates {
		e := tree.SQLFunctions[expr.FunctionName]
		cols[i] = exprToField(&tree.ExpressionFunction{
			Wrapped:  false,
			Function: e,
			//Inputs:   nil,
			//Distinct: false,
		}, p.input)
	}

	return newSchema(cols...)
}

func (p *LogicalAggregate) String() string {
	groupBy := make([]string, len(p.groupBy))
	for i, expr := range p.groupBy {
		groupBy[i] = expr.ToSQL()
	}

	if len(p.aggregates) != 0 {
		aggrs := make([]string, len(p.aggregates))
		for i, expr := range p.aggregates {
			aggrs[i] = expr.ToSQL()
		}
		return fmt.Sprintf("Aggregate: groupBy=%s, aggregates=%s",
			strings.Join(groupBy, ","), strings.Join(aggrs, ","))
	}
	return fmt.Sprintf("Aggregate: groupBy=%s",
		strings.Join(groupBy, ","))
}

func (p *LogicalAggregate) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalAggregate) Accept(v LogicalVisitor) any {
	return v.VisitLogicalAggregate(p)
}

type LogicalLimit struct {
	input  LogicalPlan
	limit  tree.Expression
	offset tree.Expression
}

func NewLogicalLimit(input LogicalPlan, limit tree.Expression, offset tree.Expression) *LogicalLimit {
	return &LogicalLimit{
		input:  input,
		limit:  limit,
		offset: offset,
	}
}

func (p *LogicalLimit) Schema() *schema {
	// limit won't change schema
	return p.input.Schema()
}

func (p *LogicalLimit) String() string {
	if p.offset != nil {
		return "Limit: " + p.limit.ToSQL() + " offset " + p.offset.ToSQL()
	}
	return "Limit: " + p.limit.ToSQL()
}

func (p *LogicalLimit) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalLimit) Accept(v LogicalVisitor) any {
	return v.VisitLogicalLimit(p)
}

type LogicalSort struct {
	input   LogicalPlan
	orderBy []*tree.OrderingTerm
}

func NewLogicalTakeN(input LogicalPlan, orderBy []*tree.OrderingTerm) *LogicalSort {
	return &LogicalSort{
		input:   input,
		orderBy: orderBy,
	}
}

func (p *LogicalSort) Schema() *schema {
	// takeN won't change schema
	return p.input.Schema()
}

func (p *LogicalSort) String() string {
	orderBy := ""
	if p.orderBy != nil {
		for _, term := range p.orderBy {
			orderBy += term.ToSQL() + ","
		}
	}

	return "Sort: By=" + orderBy
}

func (p *LogicalSort) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalSort) Accept(v LogicalVisitor) any {
	return v.VisitLogicalSort(p)
}

type LogicalJoin struct {
	inputL LogicalPlan
	inputR LogicalPlan

	joinType tree.JoinType
	on       tree.Expression
}

func NewLogicalJoin(inputL LogicalPlan, inputR LogicalPlan, joinType tree.JoinType, on tree.Expression) *LogicalJoin {
	return &LogicalJoin{
		inputL:   inputL,
		inputR:   inputR,
		joinType: joinType,
		on:       on,
	}
}

func (p *LogicalJoin) Schema() *schema {
	// join should return a new schema based on the input schemas, combine columns
	panic("not implemented")
}

func (p *LogicalJoin) String() string {
	return "Join: "
	//return "Join: " + p.joinType
}

func (p *LogicalJoin) Inputs() []LogicalPlan {
	return []LogicalPlan{p.inputL, p.inputR}
}

func (p *LogicalJoin) Accept(v LogicalVisitor) any {
	return v.VisitLogicalJoin(p)
}

type LogicalSet struct {
	inputL LogicalPlan
	inputR LogicalPlan

	setType tree.CompoundOperatorType
}

func NewLogicalSet(inputL LogicalPlan, inputR LogicalPlan, setType tree.CompoundOperatorType) *LogicalSet {
	return &LogicalSet{
		inputL:  inputL,
		inputR:  inputR,
		setType: setType,
	}
}

func (p *LogicalSet) Schema() *schema {
	// use left schema as the schema of set
	return p.inputL.Schema()
}

func (p *LogicalSet) String() string {
	return "Set: " + p.setType.ToSQL()
}

func (p *LogicalSet) Inputs() []LogicalPlan {
	return []LogicalPlan{p.inputL, p.inputR}
}

func (p *LogicalSet) Accept(v LogicalVisitor) any {
	return v.VisitLogicalSet(p)
}

type LogicalDistinct struct {
	input LogicalPlan
	//keys  []tree.Expression
	keys []tree.ResultColumn
}

func NewLogicalDistinct(input LogicalPlan, keys []tree.ResultColumn) *LogicalDistinct {
	return &LogicalDistinct{
		input: input,
		keys:  keys,
	}
}

func (p *LogicalDistinct) Schema() *schema {
	// distinct won't change schema
	return p.input.Schema()
}

func (p *LogicalDistinct) String() string {
	return "Distinct: "
	//return "Distinct: " + strings.Join(p.keys, ",")
}

func (p *LogicalDistinct) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalDistinct) Accept(v LogicalVisitor) any {
	return v.VisitLogicalDistinct(p)
}
