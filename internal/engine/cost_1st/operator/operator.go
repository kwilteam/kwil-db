package operator

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

type OperatorType uint16

const (
	unknownOp OperatorType = iota
	// Logical operator
	Logical
	LogicalSeqScan
	LogicalFilter
	LogicalLimit
	LogicalTakeN
	LogicalJoin
	LogicalSet
	LogicalUnion
	LogicalIntersect
	LogicalExcept
	LogicalAggr
	LogicalDistinct
	LogicalProjection
	// Physical operator
	//Physical
)

var operatorName = [...]string{
	unknownOp: "unknown",
	// Logical operator
	Logical:          "Logical",
	LogicalSeqScan:   "LogicalSeqScan",
	LogicalFilter:    "LogicalFilter",
	LogicalLimit:     "LogicalLimit",
	LogicalTakeN:     "LogicalTakeN",
	LogicalJoin:      "LogicalJoin",
	LogicalSet:       "LogicalSet",
	LogicalUnion:     "LogicalUnion",
	LogicalIntersect: "LogicalIntersect",
	LogicalExcept:    "LogicalExcept",
	LogicalAggr:      "LogicalAggr",
	// Physical operator
}

type OutputColumn struct {
	OriginalTblName string
	OriginalColName string
	TblName         string
	ColName         string
	DB              string

	used bool
}

type Outputs []*OutputColumn

type Operator interface {
	//// Outputs returns the output names of each column.
	//Outputs() []*OutputColumn
	//// SetOutputs sets the output name by the given slice.
	//SetOutputs([]*OutputColumn)
	//// Schema returns the schema of the operator.
	//Schema() *schema
	//// SetSchema sets the schema for the operator.
	//SetSchema(*schema)

	fmt.Stringer
	// Accept accepts an visitor to visit itself.
	Accept(Visitor) any
	// OpType returns the operator type.
	OpType() OperatorType
}

type schema struct {
	cols []*types.Column
	keys []*types.Index
}

// baseOperator is the base struct for all operator nodes.
// It implements the Operator interface except Accept/String/OpType methods.
type baseOperator struct {
	schema  *schema
	outputs []*OutputColumn
}

//
//func (n *baseOperator) Schema() *schema {
//	return n.schema
//}
//
//func (n *baseOperator) SetSchema(schema *schema) {
//	n.schema = schema
//}
//
//func (n *baseOperator) Outputs() []*OutputColumn {
//	return n.outputs
//}
//
//func (n *baseOperator) SetOutputs(outputs []*OutputColumn) {
//	n.outputs = outputs
//}

// LogicalScanOperator represents a logical scan operator.
type LogicalScanOperator struct {
	baseOperator

	Table      string
	Alias      string
	OutputCols []tree.ResultColumn
}

func NewLogicalScanOperator(table string, alias string, outputCols []tree.ResultColumn) *LogicalScanOperator {
	return &LogicalScanOperator{
		Table:      table,
		Alias:      alias,
		OutputCols: outputCols,
	}
}

func (op *LogicalScanOperator) OpType() OperatorType {
	return LogicalSeqScan
}

func (op *LogicalScanOperator) String() string {
	outputCols := op.OutputCols[0].ToSQL()
	for _, col := range op.OutputCols[1:] {
		outputCols += ", " + col.ToSQL()
	}

	if op.Alias != "" {
		return fmt.Sprintf("LogicalScan table=%s(%s) columns=%s", op.Table, op.Alias, outputCols)
	}
	return fmt.Sprintf("LogicalScan table=%s columns=%s", op.Table, outputCols)
}

func (op *LogicalScanOperator) Accept(v Visitor) any {
	return v.VisitLogicalScan(op)
}

// LogicalTakeNOperator represents a logical orderBy & limit operator.
// TODO: probably should merge with LogicalLimitOperator
type LogicalTakeNOperator struct {
	baseOperator

	limit  tree.Expression // NOTE: it's expr in tree.Limit, so this is realized
	offset tree.Expression

	orderBy *tree.OrderBy // TODO: the expression should be a column reference(or resolver)
}

func NewLogicalTakeNOperator(by *tree.OrderBy,
	limit tree.Expression, offset tree.Expression) *LogicalTakeNOperator {
	return &LogicalTakeNOperator{
		// ignore limit & offset for now
		//limit:   limit,
		//offset:  offset,

		orderBy: by,
	}
}

func (op *LogicalTakeNOperator) Accept(v Visitor) any {
	return v.VisitLogicalTakeN(op)
}

func (op *LogicalTakeNOperator) OpType() OperatorType {
	return LogicalTakeN
}

func (op *LogicalTakeNOperator) String() string {
	orderBy := ""
	if op.orderBy != nil {
		orderBy = op.orderBy.ToSQL()
	}

	// TODO: should be realized value
	if op.limit != nil {
		if op.offset != nil {
			return fmt.Sprintf(
				"LogicalTakeN %s limit=%s offset=%s",
				orderBy, op.limit.ToSQL(), op.offset.ToSQL())
		}
		return fmt.Sprintf("LogicalTakeN %s limit=%s",
			orderBy, op.limit.ToSQL())
	}

	return fmt.Sprintf("LogicalTakeN %s", orderBy)
}

// LogicalLimitOperator represents a logical limit operator.
type LogicalLimitOperator struct {
	baseOperator

	limit  tree.Expression // TODO: should be realized value
	offset tree.Expression
}

func NewLogicalLimitOperator(limit tree.Expression, offset tree.Expression) *LogicalLimitOperator {
	return &LogicalLimitOperator{
		limit:  limit,
		offset: offset,
	}
}

func (op *LogicalLimitOperator) Accept(v Visitor) any {
	return v.VisitLogicalLimit(op)
}

func (op *LogicalLimitOperator) OpType() OperatorType {
	return LogicalLimit
}

func (op *LogicalLimitOperator) String() string {
	//return fmt.Sprintf("LogicalLimit limit=%d offset=%d\op", op.limit, op.offset)
	// TODO: should be realized value
	if op.offset != nil {
		return fmt.Sprintf("LogicalLimit limit=%s offset=%s", op.limit.ToSQL(), op.offset.ToSQL())
	}
	return fmt.Sprintf("LogicalLimit limit=%s", op.limit.ToSQL())
}

// LogicalFilterOperator represents a logical filter operator.
type LogicalFilterOperator struct {
	baseOperator

	filter string
}

func NewLogicalFilterOperator(filter string) *LogicalFilterOperator {
	return &LogicalFilterOperator{
		filter: filter,
	}
}

func (op *LogicalFilterOperator) Accept(v Visitor) any {
	return v.VisitLogicalFilter(op)
}

func (op *LogicalFilterOperator) OpType() OperatorType {
	return LogicalFilter
}

func (op *LogicalFilterOperator) String() string {
	return fmt.Sprintf("LogicalFilter filter=%s", op.filter)
}

// LogicalJoinOperator represents a logical join operator.
type LogicalJoinOperator struct {
	baseOperator

	joinOP *tree.JoinOperator
	on     string
}

func NewLogicalJoinOperator(join *tree.JoinOperator, on string) *LogicalJoinOperator {
	return &LogicalJoinOperator{
		joinOP: join,
		on:     on,
	}
}

func (op *LogicalJoinOperator) Accept(v Visitor) any {
	//switch n.joinType {
	//case tree.JoinTypeInner:
	//
	//}

	return v.VisitLogicalJoin(op)
	//return nil
}

func (op *LogicalJoinOperator) OpType() OperatorType {
	return LogicalJoin
}

func (op *LogicalJoinOperator) String() string {
	return fmt.Sprintf("LogicalJoin %s on=%s", op.joinOP.ToSQL(), op.on)
}

//type ILogicalSetOperator interface {
//	setOperator()
//}

type LogicalSetOperator struct {
	baseOperator

	setType tree.CompoundOperatorType
}

func NewLogicalSetOperator(setType tree.CompoundOperatorType) *LogicalSetOperator {
	return &LogicalSetOperator{
		setType: setType,
	}
}

func (op *LogicalSetOperator) Accept(v Visitor) any {
	//// or do this dispatch in visitor?
	//switch op.setType {
	//case tree.CompoundOperatorTypeUnion:
	//	return v.VisitLogicalUnion(op)
	////case tree.CompoundOperatorTypeUnionAll:
	////	return v.VisitLogicalUnionAll(op)
	//case tree.CompoundOperatorTypeIntersect:
	//	return v.VisitLogicalIntersect(op)
	//case tree.CompoundOperatorTypeExcept:
	//	return v.VisitLogicalExcept(op)
	//default:
	//	panic(fmt.Errorf("unknown compound operator type %d", op.setType))
	//}

	return v.VisitLogicalSet(op)

}

func (op *LogicalSetOperator) OpType() OperatorType {
	return LogicalSet
}

func (op *LogicalSetOperator) String() string {
	return fmt.Sprintf("LogicalSet setType=%s", op.setType)
}

//func (n *LogicalSetOperator) setOperator() {}

// LogicalAggregateOperator represents a logical aggregrate/groupby operator.
type LogicalAggregateOperator struct {
	baseOperator

	keys []tree.Expression // tree.ExpressionColumn
	// aggregate func -> col
	colsAggrFuncs map[*tree.AggregateFunc]*tree.ExpressionColumn
}

func NewLogicalAggregateOperator(keys []tree.Expression,
	colAggrFuncs map[*tree.AggregateFunc]*tree.ExpressionColumn) *LogicalAggregateOperator {
	return &LogicalAggregateOperator{
		keys:          keys,
		colsAggrFuncs: colAggrFuncs,
	}
}

func (op *LogicalAggregateOperator) Accept(v Visitor) any {
	//switch op.aggType {
	//case tree.SelectTypeDistinct:
	//	return v.VisitLogicalDistinct(op)
	//case tree.SelectTypeAll:
	//	return v.VisitLogicalAll(op)
	//default:
	//	panic(fmt.Errorf("unknown select type %d", op.aggType))
	//}

	return v.VisitLogicalAggregate(op)
}

func (op *LogicalAggregateOperator) OpType() OperatorType {
	return LogicalAggr
}

func (op *LogicalAggregateOperator) String() string {
	bys := make([]string, 0, len(op.keys))
	for _, col := range op.keys {
		bys = append(bys, col.ToSQL())
	}

	return fmt.Sprintf("LogicalAggregate by=%s", strings.Join(bys, ","))
}

type LogicalDistinctOperator struct {
	baseOperator

	keys []tree.ResultColumn
}

func NewLogicalDistinctOperator(keys []tree.ResultColumn) *LogicalDistinctOperator {
	return &LogicalDistinctOperator{
		keys: keys,
	}
}

func (op *LogicalDistinctOperator) Accept(v Visitor) any {
	return v.VisitLogicalDistinct(op)
}

func (op *LogicalDistinctOperator) OpType() OperatorType {
	return LogicalDistinct
}

func (op *LogicalDistinctOperator) String() string {
	bys := make([]string, 0, len(op.keys))
	for _, col := range op.keys {
		bys = append(bys, col.ToSQL())
	}

	return fmt.Sprintf("LogicalDistinct by=%s", strings.Join(bys, ","))
}

type LogicalProjectionOperator struct {
	baseOperator

	projections []tree.Expression
}

func NewLogicalProjectionOperator(projections []tree.Expression) *LogicalProjectionOperator {
	return &LogicalProjectionOperator{
		projections: projections,
	}
}

func (op *LogicalProjectionOperator) Accept(v Visitor) any {
	return v.VisitLogicalProjection(op)
}

func (op *LogicalProjectionOperator) OpType() OperatorType {
	return LogicalProjection
}

func (op *LogicalProjectionOperator) String() string {
	projs := make([]string, 0, len(op.projections))
	for _, col := range op.projections {
		projs = append(projs, col.ToSQL())
	}

	return fmt.Sprintf("LogicalProjection projections=%s", strings.Join(projs, ","))
}
