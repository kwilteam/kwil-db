package models

import (
	"fmt"
	types "kwil/x/sqlx/spec"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

type SQLQuery struct {
	Name   string        `json:"name"`
	Type   string        `json:"type"`
	Table  string        `json:"table"`
	Params []Param       `json:"params,omitempty"`
	Where  []WhereClause `json:"where,omitempty"`
}

func (q *SQLQuery) Validate(db *Database) error {
	// check if query name is valid
	err := CheckName(q.Name, types.QUERY)
	if err != nil {
		return fmt.Errorf(`invalid name for query: %w`, err)
	}

	// check if table exists
	table := db.GetTable(q.Table)
	if table == nil {
		return fmt.Errorf(`table "%s" does not exist`, q.Table)
	}

	// check if query type is valid
	err = q.validateQueryType()
	if err != nil {
		return err
	}

	paramMap := make(map[string]struct{})
	for _, param := range q.Params {
		// check if parameter name is unique
		if _, ok := paramMap[param.Column]; ok {
			return fmt.Errorf(`duplicate parameter column "%s"`, param.Column)
		}
		paramMap[param.Column] = struct{}{}

		// check if parameter is valid
		err := param.Validate(table)
		if err != nil {
			return fmt.Errorf(`invalid parameter for query "%s": %w`, q.Name, err)
		}
	}

	whereMap := make(map[string]struct{})
	for _, where := range q.Where {
		// check if where column is unique
		if _, ok := whereMap[where.Column]; ok {
			return fmt.Errorf(`duplicate where column "%s"`, where.Column)
		}
		whereMap[where.Column] = struct{}{}

		// check if where is valid
		err := where.Validate(table)
		if err != nil {
			return fmt.Errorf(`invalid where for query "%s": %w`, q.Name, err)
		}
	}

	return nil
}

// Checks that the query type is valid and that the query has an
// acceptable parameters and where clauses.
func (q *SQLQuery) validateQueryType() error {
	qtype, err := types.Validation.ConvertQueryType(q.Type)
	if err != nil {
		return fmt.Errorf(`invalid type for query "%s": %w`, q.Name, err)
	}

	// check if query has too many params or where clauses
	if len(q.Params) > types.MAX_PARAM_COUNT || len(q.Where) > types.MAX_WHERE_COUNT {
		return fmt.Errorf(`query "%s" cannot have more than %d parameters or %d where clauses`, q.Name, types.MAX_PARAM_COUNT, types.MAX_WHERE_COUNT)
	}

	// check if insert has where clause
	if qtype == types.INSERT && len(q.Where) > 0 {
		return fmt.Errorf(`insert query "%s" cannot have where clause`, q.Name)
	}

	// if parameter is an insert or update, must have at least one parameter
	if qtype == types.INSERT || qtype == types.UPDATE {
		if len(q.Params) == 0 {
			return fmt.Errorf(`query "%s" must have at least one parameter`, q.Name)
		}
	}

	// check that update and delete need at least one where clause
	if qtype == types.UPDATE || qtype == types.DELETE {
		if len(q.Where) == 0 {
			return fmt.Errorf(`query "%s" must have at least one where clause`, q.Name)
		}
	}

	// delete queries cannot have parameters
	if qtype == types.DELETE && len(q.Params) > 0 {
		return fmt.Errorf(`delete query "%s" cannot have parameters`, q.Name)
	}

	return nil
}

func (q *SQLQuery) Prepare(schemaName string) (*ExecutableQuery, error) {
	switch q.Type {
	case "insert":
		return q.prepareInsert(schemaName)
	case "update":
		//return q.prepareUpdate(schemaName)
	case "delete":
		//return q.prepareDelete(schemaName)
	default:
		return nil, fmt.Errorf(`invalid query type "%s"`, q.Type)
	}

	return nil, nil
}

func (q *SQLQuery) prepareInsert(schemaName string) (*ExecutableQuery, error) {
	// sort the parameters by whether or not they are static
	// we want the non-static ones first, since the user will be filling those
	params, wheres := q.sortStatic()

	// building the statement
	stmt, err := q.buildInsertStatement(schemaName, params)
	if err != nil {
		return nil, err
	}

	return &ExecutableQuery{
		Name:         q.Name,
		Statement:    stmt,
		Type:         q.Type,
		Table:        q.Table,
		Inputs:       params,
		WhereClauses: wheres,
	}, nil
}

func (q *SQLQuery) buildInsertStatement(schemaName string, params map[int]Param) (string, error) {
	// determine table name
	var tblName string
	if schemaName == "" {
		tblName = q.Table
	} else {
		tblName = schemaName + "." + q.Table
	}

	// get the columns and values
	var cols []any
	vals := make([]any, 0, len(params))
	for _, param := range params {
		cols = append(cols, param.Column)
		vals = append(vals, struct{}{})
	}
	stmt, _, err := goqu.Dialect("postgres").Insert(tblName).Prepared(true).Cols(cols...).Vals(vals).ToSQL()
	if err != nil {
		return "", fmt.Errorf(`error preparing insert statement for query "%s": %w`, q.Name, err)
	}

	return stmt, nil
}

func (q *SQLQuery) sortStatic() (map[int]Param, map[int]WhereClause) {
	var staticParams []*Param
	var staticWhere []*WhereClause
	params := make(map[int]Param)
	where := make(map[int]WhereClause)
	i := 0
	// sorting parameters
	for _, param := range q.Params {
		if !param.Static {
			params[i] = param
		} else {
			staticParams = append(staticParams, &param)
			i++
		}
	}

	for _, staticParam := range staticParams {
		params[i] = *staticParam
		i++
	}

	// sorting where clauses
	i = 0
	for _, whereClause := range q.Where {
		if !whereClause.Static {
			where[i] = whereClause
		} else {
			staticWhere = append(staticWhere, &whereClause)
			i++
		}
	}

	for _, staticWhere := range staticWhere {
		where[i] = *staticWhere
		i++
	}

	return params, where
}
