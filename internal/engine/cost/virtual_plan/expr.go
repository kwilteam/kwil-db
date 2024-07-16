package virtual_plan

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
)

type VirtualExpr interface {
	// Evaluate evaluates the expression with the given row.
	evaluate(col datasource.Row) datasource.ColumnValue
	// Resolve returns the field that this expression represents from the input.
	Resolve(input VirtualPlan) string
}

type VLiteralStringExpr struct {
	Value string
}

func (e *VLiteralStringExpr) Resolve(_ VirtualPlan) string {
	return e.Value
}

func (e *VLiteralStringExpr) evaluate(row datasource.Row) datasource.ColumnValue {
	return datasource.NewLiteralColumnValue(e.Value)
}

type VLiteralNumericExpr struct {
	Value int64
}

func (e *VLiteralNumericExpr) Resolve(_ VirtualPlan) string {
	return fmt.Sprintf("%d", e.Value)
}

func (e *VLiteralNumericExpr) evaluate(row datasource.Row) datasource.ColumnValue {
	return datasource.NewLiteralColumnValue(e.Value)
}

type VColumnExpr struct {
	idx int
}

func (e *VColumnExpr) Resolve(plan VirtualPlan) string {
	return fmt.Sprintf("%s@%d", plan.Schema().Fields[e.idx].Name, e.idx)
}

func (e *VColumnExpr) evaluate(row datasource.Row) datasource.ColumnValue {
	return row[e.idx]
}

func compare(op string, a, b datasource.ColumnValue) bool {
	if a.Type() != b.Type() {
		panic(fmt.Sprintf("cannot compare %s and %s", a.Type(), b.Type()))
	}

	switch op {
	case "AND":
		return a.Value().(bool) && b.Value().(bool)
	case "OR":
		return a.Value().(bool) || b.Value().(bool)
	case "=":
		return a.Value() == b.Value()
	case "!=", "<>":
		return a.Value() != b.Value()
	case ">":
		return a.Value().(int) > b.Value().(int)
	case "<":
		return a.Value().(int) < b.Value().(int)
	case ">=":
		return a.Value().(int) >= b.Value().(int)
	case "<=":
		return a.Value().(int) <= b.Value().(int)
	default:
		panic(fmt.Sprintf("unknown operator %s", op))
	}
}

func VColumn(idx int) VirtualExpr {
	return &VColumnExpr{idx: idx}
}

type VBoolUnaryExpr struct {
	expr VirtualExpr
	op   string
}

func (e *VBoolUnaryExpr) Resolve(input VirtualPlan) string {
	return fmt.Sprintf("%s %s", e.op, e.expr.Resolve(input))
}

func (e *VBoolUnaryExpr) evaluate(row datasource.Row) datasource.ColumnValue {
	val := e.expr.evaluate(row)
	switch e.op {
	case "NOT":
		return datasource.NewLiteralColumnValue(!val.Value().(bool))
	default:
		panic(fmt.Sprintf("unknown operator %s", e.op))
	}
}

type VBoolBinaryExpr struct {
	left  VirtualExpr
	right VirtualExpr
	op    string
}

func (e *VBoolBinaryExpr) Resolve(input VirtualPlan) string {
	return fmt.Sprintf("%s %s %s", e.left.Resolve(input), e.op, e.right.Resolve(input))
}

func (e *VBoolBinaryExpr) evaluate(row datasource.Row) datasource.ColumnValue {
	left := e.left.evaluate(row)
	right := e.right.evaluate(row)
	return datasource.NewLiteralColumnValue(compare(e.op, left, right))
}

func VAnd(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VBoolBinaryExpr{left: left, right: right, op: "AND"}
}

func VOr(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VBoolBinaryExpr{left: left, right: right, op: "OR"}
}

func VEq(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VBoolBinaryExpr{left: left, right: right, op: "="}
}

func VNeq(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VBoolBinaryExpr{left: left, right: right, op: "!="}
}

func VGt(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VBoolBinaryExpr{left: left, right: right, op: ">"}
}

func VLt(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VBoolBinaryExpr{left: left, right: right, op: "<"}
}

func VGte(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VBoolBinaryExpr{left: left, right: right, op: ">="}
}

func VLte(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VBoolBinaryExpr{left: left, right: right, op: "<="}
}

type VArithmeticBinaryExpr struct {
	left  VirtualExpr
	right VirtualExpr
	op    string
}

func (e *VArithmeticBinaryExpr) Resolve(input VirtualPlan) string {
	return fmt.Sprintf("%s %s %s", e.left.Resolve(input), e.op, e.right.Resolve(input))
}

func (e *VArithmeticBinaryExpr) evaluate(row datasource.Row) datasource.ColumnValue {
	left := e.left.evaluate(row)
	right := e.right.evaluate(row)

	switch e.op {
	case "+":
		return datasource.NewLiteralColumnValue(left.Value().(int) + right.Value().(int))
	case "-":
		return datasource.NewLiteralColumnValue(left.Value().(int) - right.Value().(int))
	case "*":
		return datasource.NewLiteralColumnValue(left.Value().(int) * right.Value().(int))
	case "/":
		return datasource.NewLiteralColumnValue(left.Value().(int) / right.Value().(int))
	default:
		panic(fmt.Sprintf("unknown operator %s", e.op))
	}
}

func VAdd(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VArithmeticBinaryExpr{left: left, right: right, op: "+"}
}

func VSub(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VArithmeticBinaryExpr{left: left, right: right, op: "-"}
}

func VMul(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VArithmeticBinaryExpr{left: left, right: right, op: "*"}
}

func VDiv(left VirtualExpr, right VirtualExpr) VirtualExpr {
	return &VArithmeticBinaryExpr{left: left, right: right, op: "/"}
}

//type VAggregateExpr struct {
//	expr VirtualExpr
//	op   string
//}
//
//func (e *VAggregateExpr) evaluate(row datasource.Row) datasource.ColumnValue {
//
//}

type VSortExprSTUB struct {
	exprs []VirtualExpr
}

func (e *VSortExprSTUB) Resolve(input VirtualPlan) string {
	var fields []string
	for _, expr := range e.exprs {
		fields = append(fields, expr.Resolve(input))
	}
	return fmt.Sprintf("sort(%s)", fields)
}

func (e *VSortExprSTUB) evaluate(row datasource.Row) datasource.ColumnValue {
	return datasource.NewLiteralColumnValue("TODO")
}

func VSortExpr(exprs ...VirtualExpr) VirtualExpr {
	return &VSortExprSTUB{exprs: exprs}
}
