package models

import (
	"fmt"

	types "kwil/x/sqlx"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
)

// Prepare returns an executable for the query
func (q *SQLQuery) Prepare(db *Database) (*ExecutableQuery, error) {
	params, wheres := q.sortStatic()

	var stmt string
	var err error
	switch q.Type {
	case "insert":
		stmt, err = q.buildInsertStatement(db.GetSchemaName(), params)
	case "update":
		stmt, err = q.buildUpdateStatement(db.GetSchemaName(), params, wheres)
	case "delete":
		stmt, err = q.buildDeleteStatement(db.GetSchemaName(), wheres)
	default:
		return nil, fmt.Errorf(`invalid query type "%s"`, q.Type)
	}
	if err != nil {
		return nil, err
	}

	args, err := q.buildArgs(db, params, wheres)
	if err != nil {
		return nil, err
	}

	return &ExecutableQuery{
		Name:       q.Name,
		Statement:  stmt,
		Type:       q.Type,
		Table:      q.Table,
		Args:       args,
		UserInputs: q.buildUserInputs(args),
	}, nil
}

// buildInsertStatement builds the insert statement for the query
func (q *SQLQuery) buildInsertStatement(schemaName string, params []*Param) (string, error) {
	// determine table name
	var tblName string
	if schemaName == "" {
		tblName = q.Table
	} else {
		tblName = schemaName + "." + q.Table
	}

	// get the columns and values
	var cols []any
	var vals []any
	for _, param := range params {
		cols = append(cols, param.Column)
		vals = append(vals, false) // any value will do, but it can't be a struct{}{}
	}
	stmt, _, err := goqu.Dialect("postgres").Insert(tblName).Prepared(true).Cols(cols...).Vals(vals).ToSQL()
	if err != nil {
		return "", fmt.Errorf(`error preparing insert statement for query "%s": %w`, q.Name, err)
	}

	return stmt, nil
}

// buildUpdateStatement builds the update statement for the query
func (q *SQLQuery) buildUpdateStatement(schemaName string, params []*Param, wheres []*WhereClause) (string, error) {
	// determine table name
	var tblName string
	if schemaName == "" {
		tblName = q.Table
	} else {
		tblName = schemaName + "." + q.Table
	}

	// converting the parameters to a goqu record
	rec := make(exp.Record)
	for _, param := range params {
		rec[param.Column] = false // any value will do, but it can't be a struct{}{}
	}

	// converting the where clauses to goqu expressions
	var whereArray []goqu.Expression
	for _, where := range wheres {
		exp, err := operatorToPredicate(where.Operator, where.Column)
		if err != nil {
			return "", fmt.Errorf(`error preparing update statement for query "%s": %w`, q.Name, err)
		}

		whereArray = append(whereArray, exp)
	}

	// building the statement
	stmt, _, err := goqu.Dialect("postgres").Update(tblName).Prepared(true).Set(rec).Where(whereArray...).ToSQL()
	if err != nil {
		return "", fmt.Errorf(`error preparing update statement for query "%s": %w`, q.Name, err)
	}

	return stmt, nil
}

// buildDeleteStatement builds the delete statement for the query
func (q *SQLQuery) buildDeleteStatement(schemaName string, wheres []*WhereClause) (string, error) {
	// determine table name
	var tblName string
	if schemaName == "" {
		tblName = q.Table
	} else {
		tblName = schemaName + "." + q.Table
	}

	// converting the where clauses to goqu expressions
	var whereArray []goqu.Expression
	for _, where := range wheres {
		exp, err := operatorToPredicate(where.Operator, where.Column)
		if err != nil {
			return "", fmt.Errorf(`error preparing delete statement for query "%s": %w`, q.Name, err)
		}

		whereArray = append(whereArray, exp)
	}

	// building the statement
	stmt, _, err := goqu.Dialect("postgres").Delete(tblName).Prepared(true).Where(whereArray...).ToSQL()
	if err != nil {
		return "", fmt.Errorf(`error preparing delete statement for query "%s": %w`, q.Name, err)
	}

	return stmt, nil
}

// sortStatic sorts the parameters and where clauses by whether or not they are static (default)
func (q *SQLQuery) sortStatic() ([]*Param, []*WhereClause) {

	var staticParams []*Param
	params := make([]*Param, len(q.Params))
	i := 0
	// sorting parameters
	for _, param := range q.Params {
		if !param.Static {
			params[i] = param
			i++
		} else {
			staticParams = append(staticParams, param)
		}
	}

	for _, staticParam := range staticParams {
		params[i] = staticParam
		i++
	}

	// sorting where clauses
	var staticWhere []*WhereClause
	where := make([]*WhereClause, len(q.Where))
	i = 0
	for _, whereClause := range q.Where {
		if !whereClause.Static {
			where[i] = whereClause
			i++
		} else {
			staticWhere = append(staticWhere, whereClause)
		}
	}

	for _, staticWhere := range staticWhere {
		where[i] = staticWhere
		i++
	}

	return params, where
}

func (q *SQLQuery) buildArgs(db *Database, params []*Param, wheres []*WhereClause) ([]*Arg, error) {
	var args []*Arg
	i := 0

	// getting table in order to get column types
	tbl := db.GetTable(q.Table)
	if tbl == nil {
		return nil, fmt.Errorf(`table "%s" not found`, q.Table)
	}

	// adding params
	// since they are already sorted, we can just add them in order
	for _, param := range params {
		// getting column type
		arg, err := param.buildArg(tbl, i)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		i++
	}

	// adding where clauses
	for _, where := range wheres {
		arg, err := where.buildArg(tbl, i)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		i++
	}

	return args, nil
}

func (q *SQLQuery) buildUserInputs(args []*Arg) []*UserInput {
	var inputs []*UserInput
	i := 0
	for _, arg := range args {
		if !arg.Static {
			inputs = append(inputs, arg.buildInput(i))
			i++
		}
	}

	return inputs
}

func operatorToPredicate(op string, column string) (exp.Expression, error) {
	kwilOp, err := types.Conversion.ConvertComparisonOperator(op)
	if err != nil {
		return nil, err
	}

	i := "" // Goqu doesn't always like empty interfaces{} when preparing statements but does fine with empty strings
	switch kwilOp {
	case types.EQUAL:
		return goqu.C(column).Eq(i), nil
	case types.NOT_EQUAL:
		return goqu.C(column).Neq(i), nil
	case types.GREATER_THAN:
		return goqu.C(column).Gt(i), nil
	case types.GREATER_THAN_OR_EQUAL:
		return goqu.C(column).Gte(i), nil
	case types.LESS_THAN:
		return goqu.C(column).Lt(i), nil
	case types.LESS_THAN_OR_EQUAL:
		return goqu.C(column).Lte(i), nil
	}

	return nil, fmt.Errorf("unknown operator: %s", op)
}
