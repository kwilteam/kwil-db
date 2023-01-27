package validator

import (
	"fmt"
	"kwil/x/execution"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

func (v *Validator) validateQueries() error {
	queryNames := make(map[string]struct{})
	for _, q := range v.db.SQLQueries {
		// validate query name is unique
		if _, ok := queryNames[q.Name]; ok {
			return fmt.Errorf(`duplicate query name "%s"`, q.Name)
		}
		queryNames[q.Name] = struct{}{}

		err := v.validateQuery(q)
		if err != nil {
			return fmt.Errorf(`error on query %v: %w`, q.Name, err)
		}
	}

	return nil
}

func (v *Validator) validateQuery(q *databases.SQLQuery[anytype.KwilAny]) error {
	if !q.Type.IsValid() {
		return fmt.Errorf(`invalid query type`)
	}

	// check if query name is valid
	err := CheckName(q.Name, databases.MAX_QUERY_NAME_LENGTH)
	if err != nil {
		return fmt.Errorf(`invalid name for query: %w`, err)
	}

	// check if table exists
	table := v.db.GetTable(q.Table)
	if table == nil {
		return fmt.Errorf(`table "%s" does not exist`, q.Table)
	}

	// check if query type is valid
	err = validateQueryType(q)
	if err != nil {
		return err
	}

	paramColumns := make(map[string]struct{}) // for guaranteeing that each column is only used at most once
	inputNames := make(map[string]struct{})   // for guaranteeing that each name is only used at most once for both params and where
	for _, param := range q.Params {
		// check if parameter name is unique
		if _, ok := inputNames[param.Name]; ok {
			return fmt.Errorf(`duplicate parameter name "%s"`, param.Name)
		}
		inputNames[param.Name] = struct{}{}

		// check if parameter column is unique
		if _, ok := paramColumns[param.Column]; ok {
			return fmt.Errorf(`duplicate parameter column "%s"`, param.Column)
		}
		paramColumns[param.Column] = struct{}{}

		// check if parameter is valid
		err := validateInput(param, table)
		if err != nil {
			return fmt.Errorf(`invalid parameter for param "%s": %w`, param.Name, err)
		}
	}
	for _, where := range q.Where {
		// check if where name is unique
		if _, ok := inputNames[where.Name]; ok {
			return fmt.Errorf(`duplicate where name "%s"`, where.Name)
		}
		inputNames[where.Name] = struct{}{}

		// check operator
		if !where.Operator.IsValid() {
			return fmt.Errorf(`unknown operator: %d`, where.Operator.Int())
		}

		// check if where is valid
		err := validateInput(where, table)
		if err != nil {
			return fmt.Errorf(`invalid where for where_clause "%s": %w`, where.Name, err)
		}
	}

	return nil
}

// Checks that the query type is valid and that the query has an
// acceptable parameters and where clauses.
func validateQueryType(q *databases.SQLQuery[anytype.KwilAny]) error {
	// check if query has too many params or where clauses
	if len(q.Params) > execution.MAX_PARAM_PER_QUERY || len(q.Where) > execution.MAX_WHERE_PER_QUERY {
		return fmt.Errorf(`query "%s" cannot have more than %d parameters or %d where clauses`, q.Name, execution.MAX_PARAM_PER_QUERY, execution.MAX_WHERE_PER_QUERY)
	}

	// check if insert has where clause
	if q.Type == execution.INSERT && len(q.Where) > 0 {
		return fmt.Errorf(`insert query "%s" cannot have where clause`, q.Name)
	}

	// if parameter is an insert or update, must have at least one parameter
	if q.Type == execution.INSERT || q.Type == execution.UPDATE {
		if len(q.Params) == 0 {
			return fmt.Errorf(`query "%s" must have at least one parameter`, q.Name)
		}
	}

	// check that update and delete need at least one where clause
	if q.Type == execution.UPDATE || q.Type == execution.DELETE {
		if len(q.Where) == 0 {
			return fmt.Errorf(`query "%s" must have at least one where clause`, q.Name)
		}
	}

	// delete queries cannot have parameters
	if q.Type == execution.DELETE && len(q.Params) > 0 {
		return fmt.Errorf(`delete query "%s" cannot have parameters`, q.Name)
	}

	return nil
}

func validateInput(input databases.Input[anytype.KwilAny], table *databases.Table[anytype.KwilAny]) error {
	// check if column exists
	col := table.GetColumn(input.GetColumn())
	if col == nil {
		return fmt.Errorf(`column "%s" does not exist`, input.GetColumn())
	}

	// check that modifier is valid
	if !input.GetModifier().IsValid() {
		return fmt.Errorf(`unknown modifier: %d`, input.GetModifier().Int())
	}

	if input.GetStatic() {

		// check if value is set
		if input.GetValue().Value() == nil && input.GetModifier() != execution.CALLER {
			return fmt.Errorf(`value must be set for non-fillable parameter on column "%s"`, input.GetColumn())
		}

		// check if value type matches column type
		if col.Type != input.GetValue().Type() && input.GetModifier() != execution.CALLER {
			return fmt.Errorf(`value "%s" must be of type "%s" for parameter on column "%s"`, fmt.Sprint(input.GetValue()), col.Type.String(), input.GetColumn())
		}

	} else { // not static: users can fill in the value
		if input.GetValue().Bytes() != nil {
			return fmt.Errorf(`value must not be set for fillable parameter on column "%s"`, input.GetColumn())
		}

		if input.GetModifier() == execution.CALLER {
			return fmt.Errorf(`modifier must not be caller for fillable parameter on column "%s"`, input.GetColumn())
		}
	}

	return nil
}
