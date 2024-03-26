package join

import "github.com/kwilteam/kwil-db/parse/sql/tree"

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

The joinAnalyzer is used in a DFS manner to determine if a join is valid.  It can exist in one of the following states:
- joinableStatusInvalid: the join is invalid
- joinableStatusContainsColumn: the join contains a column from one of the tables
- joinableStatusValid: the join is valid

*/

// checkJoin is used to check if a join is valid
func checkJoin(join *tree.JoinPredicate) error {
	if validateJoinStatus(join.Constraint) != joinableStatusValid {
		return ErrInvalidJoinCondition
	}

	return nil
}

func validateJoinStatus(joinConstraint tree.Expression) joinableStatus {
	switch j := joinConstraint.(type) {
	case *tree.ExpressionTextLiteral, *tree.ExpressionNumericLiteral, *tree.ExpressionBooleanLiteral,
		*tree.ExpressionNullLiteral, *tree.ExpressionBlobLiteral:
		return joinableStatusInvalid
	case *tree.ExpressionBindParameter:
		return joinableStatusInvalid
	case *tree.ExpressionColumn:
		return joinableStatusContainsColumn
	case *tree.ExpressionUnary:
		return joinableStatusInvalid
	case *tree.ExpressionBinaryComparison:
		left := validateJoinStatus(j.Left)
		right := validateJoinStatus(j.Right)
		if left == joinableStatusContainsColumn && right == joinableStatusContainsColumn && j.Operator == tree.ComparisonOperatorEqual {
			return joinableStatusValid
		}

		if left == joinableStatusContainsColumn || right == joinableStatusContainsColumn {
			return joinableStatusContainsColumn
		}

		return joinableStatusInvalid
	case *tree.ExpressionFunction:
		return joinableStatusInvalid
	case *tree.ExpressionList:
		return joinableStatusInvalid
	case *tree.ExpressionCollate:
		return validateJoinStatus(j.Expression)
	case *tree.ExpressionStringCompare:
		if validateJoinStatus(j.Left) == joinableStatusContainsColumn || validateJoinStatus(j.Right) == joinableStatusContainsColumn {
			return joinableStatusContainsColumn
		}

		return joinableStatusInvalid
	case *tree.ExpressionIs:
		if validateJoinStatus(j.Left) == joinableStatusContainsColumn || validateJoinStatus(j.Right) == joinableStatusContainsColumn {
			return joinableStatusContainsColumn
		}

		return joinableStatusInvalid
	case *tree.ExpressionBetween:
		return validateJoinStatus(j.Expression)
	case *tree.ExpressionSelect:
		return joinableStatusInvalid
	case *tree.ExpressionCase:
		return joinableStatusInvalid
	case *tree.ExpressionArithmetic:
		if validateJoinStatus(j.Left) == joinableStatusContainsColumn || validateJoinStatus(j.Right) == joinableStatusContainsColumn {
			return joinableStatusContainsColumn
		}

		return joinableStatusInvalid
	}

	return joinableStatusInvalid
}

type joinableStatus uint8

const (
	joinableStatusInvalid joinableStatus = iota
	joinableStatusContainsColumn
	joinableStatusValid
)
