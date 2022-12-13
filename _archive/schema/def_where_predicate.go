package schema

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type Operator int

const (
	Equal Operator = iota
	NotEqual
	GreaterThan
	GreaterThanOrEqual
	LessThan
	LessThanOrEqual
	Like
	NotLike
)

var (
	WhereConversions = &whereConversions{}
)

type whereConversions struct{}

func (w *whereConversions) StringToOperator(op string) Operator {
	switch op {
	case "=":
		return Equal
	case "!=":
		return NotEqual
	case ">":
		return GreaterThan
	case ">=":
		return GreaterThanOrEqual
	case "<":
		return LessThan
	case "<=":
		return LessThanOrEqual
	case "like":
		return Like
	case "not like":
		return NotLike
	}
	return -1
}

func (w *whereConversions) OperatorToPredicate(op Operator, column string) exp.Expression {
	i := "" // Goqu doesn't always like empty interfaces{} when preparing statements but does fine with empty strings
	switch op {
	case Equal:
		return goqu.C(column).Eq(i)
	case NotEqual:
		return goqu.C(column).Neq(i)
	case GreaterThan:
		return goqu.C(column).Gt(i)
	case GreaterThanOrEqual:
		return goqu.C(column).Gte(i)
	case LessThan:
		return goqu.C(column).Lt(i)
	case LessThanOrEqual:
		return goqu.C(column).Lte(i)
	case Like:
		return goqu.C(column).Like(i)
	case NotLike:
		return goqu.C(column).NotLike(i)
	}
	return nil
}

func (w *whereConversions) StringToPredicate(op string, column string) exp.Expression {
	return w.OperatorToPredicate(w.StringToOperator(op), column)
}
