package cost_2nd

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

type LogicalScan struct {
	basePlan

	ds DataSource // NOTE: maybe no need, just use schema a field to initialize LogicalScan
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

	if p.ds.Schema().tblAlias != "" {
		return fmt.Sprintf("Scan: table=%s(%s)", p.ds.Schema().tblName, p.ds.Schema().tblAlias)
	}

	return fmt.Sprintf("Scan: table=%s", p.ds.Schema().tblName)
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
	basePlan

	input LogicalPlan
	exprs []*tree.ResultColumnExpression
}

func NewLogicalProjection(input LogicalPlan, exprs []*tree.ResultColumnExpression) *LogicalProjection {
	return &LogicalProjection{
		input: input,
		exprs: exprs,
	}
}

func (p *LogicalProjection) Schema() *schema {
	//return Schema(expr.map { it.toField(input) })

	// projection should return a new schema based on the input schema
	//return p.input.Schema().Select(p.exprs)
	cols := make([]*field, len(p.exprs))
	for i, expr := range p.exprs {
		cols[i] = exprToField(expr.Expression, p.input)
	}
	return newSchema(cols...)
}

func (p *LogicalProjection) String() string {
	fields := make([]string, len(p.exprs))
	for i, expr := range p.exprs {
		fields[i] = removeWhitespace(expr.ToSQL())
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

type LogicalSubquery struct {
	basePlan

	input  LogicalPlan
	name   string
	schema *schema // for rename
}

func NewLogicalSubquery(input LogicalPlan, name string) *LogicalSubquery {
	plan := &LogicalSubquery{
		input:  input,
		name:   name,
		schema: input.Schema(), // NOTE
	}

	//if name != "" {
	//	plan.schema.tblAlias = name
	//	for _, field := range plan.schema.fields {
	//
	//	}
	//
	//}

	return plan
}

func (p *LogicalSubquery) Schema() *schema {
	return p.schema
}

func (p *LogicalSubquery) String() string {
	if p.name != "" {
		return "Subquery: As " + p.name
	}
	return "Subquery"
}

func (p *LogicalSubquery) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalSubquery) Accept(v LogicalVisitor) any {
	return v.VisitLogicalSubquery(p)
}

type LogicalFilter struct {
	basePlan

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
	return "Filter: " + removeWhitespace(p.expr.ToSQL())
}

func (p *LogicalFilter) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalFilter) Accept(v LogicalVisitor) any {
	return v.VisitLogicalFilter(p)
}

type LogicalAggregate struct {
	basePlan

	input LogicalPlan
	// select a,b,agg(c) from t group by a,b;
	// the output schema always contain all group by columns
	// groupBy should be 'a,b'
	groupBy []tree.Expression
	//aggregates []tree.Expression
	// aggregates should be 'agg(c)'
	aggregates []*tree.ExpressionFunction
}

func NewLogicalAggregate(input LogicalPlan, groupBy []tree.Expression,
	aggregates []*tree.ExpressionFunction) *LogicalAggregate {
	return &LogicalAggregate{
		input:      input,
		groupBy:    groupBy,
		aggregates: aggregates,
	}
}

func (p *LogicalAggregate) Schema() *schema {
	// aggregate should return a new schema based on the input schema
	//return p.input.Schema().Select(p.groupBy)
	//panic("not implemented")

	cols := make([]*field, len(p.groupBy)+len(p.aggregates))
	i := 0
	for _, expr := range p.groupBy {
		cols[i] = exprToField(expr, p.input)
		i++
	}

	for j, expr := range p.aggregates {
		cols[i+j] = exprToField(expr, p.input)
	}

	return newSchema(cols...)
}

func (p *LogicalAggregate) String() string {
	groupBy := make([]string, len(p.groupBy))
	for i, expr := range p.groupBy {
		groupBy[i] = removeWhitespace(expr.ToSQL())
	}

	if len(p.aggregates) != 0 {
		aggrs := make([]string, len(p.aggregates))
		for i, expr := range p.aggregates {
			aggrs[i] = removeWhitespace(expr.ToSQL())
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
	basePlan

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
		return "Limit: " + removeWhitespace(p.limit.ToSQL()) + " offset " + removeWhitespace(p.offset.ToSQL())
	}
	return "Limit: " + removeWhitespace(p.limit.ToSQL())
}

func (p *LogicalLimit) Inputs() []LogicalPlan {
	return []LogicalPlan{p.input}
}

func (p *LogicalLimit) Accept(v LogicalVisitor) any {
	return v.VisitLogicalLimit(p)
}

type LogicalSort struct {
	basePlan

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
			orderBy += removeWhitespace(term.ToSQL()) + ","
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
	basePlan

	inputL LogicalPlan
	inputR LogicalPlan

	joinType tree.JoinType
	on       tree.Expression // should it be a list?
}

func NewLogicalJoin(inputL LogicalPlan, inputR LogicalPlan, joinType tree.JoinType, on tree.Expression) *LogicalJoin {
	return &LogicalJoin{
		inputL:   inputL,
		inputR:   inputR,
		joinType: joinType,
		on:       on,
	}
}

// Schmea of join is the combination of left and right schema
func (p *LogicalJoin) Schema() *schema {
	// TODO: remove duplicate keys
	// if left join or inner join, remove duplicate keys from right schema
	// if right join, remove duplicate keys from left schema
	cols := make([]*field, len(p.inputL.Schema().fields)+len(p.inputR.Schema().fields))
	i := 0
	for _, field := range p.inputL.Schema().fields {
		cols[i] = field
		i++
	}

	for j, field := range p.inputR.Schema().fields {
		cols[i+j] = field
	}
	return newSchema(cols...)
}

func (p *LogicalJoin) String() string {
	return fmt.Sprintf("%s: %s", p.joinType.String(), removeWhitespace(p.on.ToSQL()))
}

func (p *LogicalJoin) Inputs() []LogicalPlan {
	return []LogicalPlan{p.inputL, p.inputR}
}

func (p *LogicalJoin) Accept(v LogicalVisitor) any {
	return v.VisitLogicalJoin(p)
}

type LogicalSet struct {
	basePlan

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
	return "Set: " + removeWhitespace(p.setType.ToSQL())
}

func (p *LogicalSet) Inputs() []LogicalPlan {
	return []LogicalPlan{p.inputL, p.inputR}
}

func (p *LogicalSet) Accept(v LogicalVisitor) any {
	return v.VisitLogicalSet(p)
}

type LogicalDistinct struct {
	basePlan

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
