package executables

import (
	"fmt"
	"kwil/pkg/databases/spec"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type wherePredicate struct {
	Column   string
	Operator spec.ComparisonOperatorType
	Value    any
}

type andSelection []wherePredicate

func (a *andSelection) asGoqu() ([]goqu.Expression, error) {
	var whereArray []goqu.Expression
	for _, where := range *a {
		exp, err := operatorToGoquExpression(where.Operator, where.Column, where.Value)
		if err != nil {
			return nil, fmt.Errorf("error converting comparison operator: %w", err)
		}

		whereArray = append(whereArray, exp)
	}

	return whereArray, nil
}

func (p *preparer) getWhereExpression() (andSelection, error) {
	var whereArray []wherePredicate

	for _, where := range p.executable.Query.Where {
		val, err := p.prepareInput(where)
		if err != nil {
			return nil, fmt.Errorf("error preparing input: %w", err)
		}

		whereArray = append(whereArray, wherePredicate{
			Column:   where.Column,
			Operator: where.Operator,
			Value:    val.Value(),
		})
	}

	return whereArray, nil
}

// Update this predicates lengths with the number of ANDs between each ORs
func (p *preparer) GetPredicateLengths() []int {
	predicateLength := make([]int, 0)
	predicateLength = append(predicateLength, len(p.executable.Query.Where))
	return predicateLength
}

func operatorToGoquExpression(op spec.ComparisonOperatorType, column string, val any) (exp.Expression, error) {
	switch op {
	case spec.EQUAL:
		return goqu.C(column).Eq(val), nil
	case spec.NOT_EQUAL:
		return goqu.C(column).Neq(val), nil
	case spec.GREATER_THAN:
		return goqu.C(column).Gt(val), nil
	case spec.GREATER_THAN_OR_EQUAL:
		return goqu.C(column).Gte(val), nil
	case spec.LESS_THAN:
		return goqu.C(column).Lt(val), nil
	case spec.LESS_THAN_OR_EQUAL:
		return goqu.C(column).Lte(val), nil
	}

	return nil, fmt.Errorf("unknown operator: %s", op.String())
}
