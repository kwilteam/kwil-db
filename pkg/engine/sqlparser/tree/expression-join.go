package tree

/*
	This file contains functionality for evaluating expressions as join conditions.
	due to SQLite's propensity for cartesian joins, we need to impose very strict standards
	on what constitutes a valid join condition.

	In the context of Kwil and SQLite, the terms "join condition" and "join constraint" are
	used interchangeably.

	As per the SQLite docs:

	`
		All joins in SQLite are based on the cartesian product of the left and right-hand datasets.
		The columns of the cartesian product dataset are, in order, all the columns of the left-hand
		dataset followed by all the columns of the right-hand dataset. There is a row in the
		cartesian product dataset formed by combining each unique combination of a row from the
		left-hand and right-hand datasets. In other words, if the left-hand dataset consists of Nleft
		rows of Mleft columns, and the right-hand dataset of Nright rows of Mright columns, then the
		cartesian product is a dataset of Nleft×Nright rows, each containing Mleft+Mright columns.

		If the join-operator is "CROSS JOIN", "INNER JOIN", "JOIN" or a comma (",") and there is no
		ON or USING clause, then the result of the join is simply the cartesian product of the left and
		right-hand datasets. If join-operator does have ON or USING clauses, those are handled according
		to the following bullet points:

		- If there is an ON clause then the ON expression is evaluated for each row of the cartesian
			product as a boolean expression. Only rows for which the expression evaluates to true
			are included from the dataset.
	`

	There are more bullet points, however this is the only one that is relevant to us.

	Due to this behavior, we need to ensure that join conditions:
	- Are binary comparisons
	- Contain columns from both sides of the join
	- Do not contain any subqueries
	- Do not contain any function calls
	- Are joined with an "=" operator

	Each side of the join condition will be evaluated recursively

	There are 3 states that an expression can be in:
	- Invalid: The expression automatically disqualifies the join
	- ContainsColumn: The expressions contains a column, and therefore can be used as one side of a join condition
	- Valid: The expression is a binary comparison where each side is 'ContainsColumn', joined with an "=", and therefore can be used as a join condition
*/

type joinable interface {
	joinable() joinableStatus
}

type joinableStatus uint8

const (
	joinableStatusInvalid joinableStatus = iota
	joinableStatusContainsColumn
	joinableStatusValid
)

func (e *ExpressionLiteral) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *ExpressionBindParameter) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *ExpressionColumn) joinable() joinableStatus {
	return joinableStatusContainsColumn
}

func (e *ExpressionUnary) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *ExpressionBinaryComparison) joinable() joinableStatus {
	if e.Left.joinable() == joinableStatusContainsColumn && e.Right.joinable() == joinableStatusContainsColumn && e.Operator == ComparisonOperatorEqual {
		return joinableStatusValid
	}

	if e.Left.joinable() == joinableStatusContainsColumn || e.Right.joinable() == joinableStatusContainsColumn {
		return joinableStatusContainsColumn
	}

	return joinableStatusInvalid
}

func (e *ExpressionFunction) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *ExpressionList) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *ExpressionCollate) joinable() joinableStatus {
	return e.Expression.joinable()
}

func (e *ExpressionStringCompare) joinable() joinableStatus {
	if e.Left.joinable() == joinableStatusContainsColumn || e.Right.joinable() == joinableStatusContainsColumn {
		return joinableStatusContainsColumn
	}

	return joinableStatusInvalid
}

func (e *ExpressionIsNull) joinable() joinableStatus {
	return e.Expression.joinable()
}

func (e *ExpressionDistinct) joinable() joinableStatus {
	if e.Left.joinable() == joinableStatusContainsColumn || e.Right.joinable() == joinableStatusContainsColumn {
		return joinableStatusContainsColumn
	}

	return joinableStatusInvalid
}

func (e *ExpressionBetween) joinable() joinableStatus {
	return e.Expression.joinable()
}

func (e *ExpressionSelect) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *ExpressionCase) joinable() joinableStatus {
	return joinableStatusInvalid
}

func (e *ExpressionArithmetic) joinable() joinableStatus {
	if e.Left.joinable() == joinableStatusContainsColumn || e.Right.joinable() == joinableStatusContainsColumn {
		return joinableStatusContainsColumn
	}

	return joinableStatusInvalid
}

func (e *ExpressionRaise) joinable() joinableStatus {
	return joinableStatusInvalid
}
