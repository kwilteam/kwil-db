package planner2

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
	"github.com/kwilteam/kwil-db/parse"
)

// LogicalNode represents a logical operation on a relation,
// which returns a new relation.
type LogicalNode interface {
	Relation
	EvaluateCost(Catalog, *EvalutationContext) *RelationStatistics
}

// Table represents a table in the database.
type Table struct {
	PGSchema string
	Name     string
}

func (t *Table) Schema(c Catalog) *Schema {
	ds, err := c.GetDataSource(t.PGSchema, t.Name)
	if err != nil {
		panic(err)
	}
	return ds.Schema()
}

func (t *Table) EvaluateCost(c Catalog, ctx *EvalutationContext) *RelationStatistics {
	// natively, no cost. Cost gets applied at the scan level.
	ds, err := c.GetDataSource(t.PGSchema, t.Name)
	if err != nil {
		panic(err)
	}

	stats := ds.Statistics()

	schema := ds.Schema()
	relStats := &RelationStatistics{
		RowCount: stats.RowCount,
	}

	relStats.ColumnStatistics = make(map[[2]string]*ColumnStatistics)
	relStats.ColumnOrder = make([][2]string, len(schema.Fields))
	for i, field := range schema.Fields {
		relStats.ColumnOrder[i] = [2]string{field.ParentRelation, field.Name}
		relStats.ColumnStatistics[[2]string{field.ParentRelation, field.Name}] = &ColumnStatistics{
			NullCount:     stats.ColumnStatistics[i].NullCount,
			Min:           stats.ColumnStatistics[i].Min,
			Max:           stats.ColumnStatistics[i].Max,
			DistinctCount: stats.ColumnStatistics[i].DistinctCount,
			AvgSize:       stats.ColumnStatistics[i].AvgSize,
		}
	}

	return relStats
}

// ScanNode represents a scan operation on a table.
// This can be on a table in the DB, a subquery, a CTE,
// or a function call that returns a table.
type ScanNode struct {
	Table  LogicalNode
	Name   string      // Name or alias of the relation being scanned.
	Filter LogicalExpr // Filter to apply to the scan.
}

func (s *ScanNode) Schema(c Catalog) *Schema {
	return s.Table.Schema(c)
}

func (s *ScanNode) EvaluateCost(c Catalog, ctx *EvalutationContext) *RelationStatistics {
	stats := s.Table.EvaluateCost(c, ctx)

	err := stats.Flatten(s.Name)
	if err != nil {
		panic(err)
	}
	schema := s.Table.Schema(c)

	// if nil, then cost is a physical scan
	if s.Filter == nil {
		ctx.Cost += scanCost(stats.RowCount, schema.RowTypes()...)
	} else {
		filterCost := s.Filter.Filter(schema, stats, ctx)
	}
}

type CompoundNode struct {
	Left     LogicalNode
	Right    LogicalNode
	Operator parse.CompoundOperator
}

func (c *CompoundNode) Schema(c2 Catalog) *Schema {
	return c.Left.Schema(c2)
}

type ProjectNode struct {
	Source LogicalNode
	Exprs  []LogicalExpr
}

func (p *ProjectNode) Schema(c Catalog) *Schema {
	f := make([]*Field, 0)
	for _, expr := range p.Exprs {
		f = append(f, expr.Project(p.Source.Schema(c))...)
	}

	return &Schema{Fields: f}
}

func (p *ProjectNode) EvaluateCost(c Catalog, ctx *EvalutationContext) *datatypes.Statistics {
	pushDownProjections := make(map[string][]string) // maps the relation name to the columns that can be pushed down
	schema := p.Source.Schema(c)
	for _, expr := range schema.Fields {
		pushDownProjections[expr.ParentRelation] = append(pushDownProjections[expr.ParentRelation], expr.Name)
	}
}

// LogicalExpr represents a scalar expression.
// This is any expression that returns a single value.
type LogicalExpr interface {
	// Project returns the fields that are used by the expression.
	Project(*Schema) []*Field
	// Filter returns the amount of rows of the data after applying the filter.
	// It will also modify the evaluation context to add the cost of performing
	// the given filter.
	// It returns a function so that callers can choose whether to evaluate the
	// expression or discard it.
	Filter(*Schema, *RelationStatistics, *EvalutationContext) FilterCost
}

type FilterCost struct {
	// Cost is the cost of applying the filter once.
	Cost int64
	// Rows is the estimated number of rows that will be returned after applying the filter.
	Rows int64
	// Sargable is true if the filter is sargable.
	Sargable bool
	// CorrelatedFields is a set of fields that are correlated with the filter.
	CorrelatedFields map[[2]string]struct{}
}

func (f *FilterCost) Merge(other FilterCost) {
	f.Cost += other.Cost
	f.Rows = other.Rows
	f.Sargable = f.Sargable && other.Sargable
	for k := range other.CorrelatedFields {
		f.CorrelatedFields[k] = struct{}{}
	}
}

func newFilterCost() FilterCost {
	return FilterCost{
		CorrelatedFields: make(map[[2]string]struct{}),
	}
}

// ColumnExpr represents a column reference.
type ColumnExpr struct {
	Relation string // can be nil if the column is unqualified
	Name     string
}

func (c *ColumnExpr) Project(s *Schema) []*Field {
	field, _, err := s.FindField(c.Relation, c.Name)
	if err != nil {
		panic(err)
	}

	return []*Field{field}
}

func (c *ColumnExpr) Filter(s *Schema, stats *RelationStatistics, ctx *EvalutationContext) FilterCost {
	f := newFilterCost()

	field, found, err := s.FindField(c.Relation, c.Name)
	if err != nil && found {
		panic(err)
	}

	// if not found, this is a correlated field
	if !found {
		f.CorrelatedFields[[2]string{c.Relation, c.Name}] = struct{}{}
		f.Rows = 1
		f.Cost = ColumnAccessCost
		return f
	}

	colStats, ok := stats.ColumnStatistics[[2]string{field.ParentRelation, field.Name}]
	if !ok {
		panic("column not found in statistics")
	}
	_ = colStats // TODO: we can provide more accurate pricing and selectivity based on column statistics
	f.Rows = stats.RowCount

	if field.HasIndex {
		f.Cost += indexSearchCost(stats.RowCount, field.Type)
		f.Sargable = true
	} else {
		f.Cost += scanCost(stats.RowCount, field.Type)
	}

	return f
}

// LiteralExpr represents a literal value.
type LiteralExpr struct {
	Value any
	Type  *types.DataType
}

func (l *LiteralExpr) Project(s *Schema) []*Field {
	return []*Field{&Field{
		Name:           "",
		ParentRelation: "",
		Type:           l.Type,
	}}
}

func (l *LiteralExpr) Filter(s *Schema, stats *RelationStatistics, ctx *EvalutationContext) FilterCost {
	f := newFilterCost()
	f.Rows = stats.RowCount
	f.Cost = 0
	return f
}

// ComparisonExpr represents a comparison between two expressions.
type ComparisonExpr struct {
	Left     LogicalExpr
	Right    LogicalExpr
	Operator parse.ComparisonOperator
}

// Project will search for the fields used in either side of the expression.
func (c *ComparisonExpr) Project(s *Schema) []*Field {
	return append(c.Left.Project(s), c.Right.Project(s)...)
}

func (c *ComparisonExpr) Filter(s *Schema, stats *RelationStatistics, ctx *EvalutationContext) FilterCost {
	f := newFilterCost()

	left := c.Left.Filter(s, stats, ctx)
	right := c.Right.Filter(s, stats, ctx)

	f.Merge(left)
	f.Merge(right)

	cost := mergeJoinCost(&left, &right)
	f.Cost += cost

	return f
}
