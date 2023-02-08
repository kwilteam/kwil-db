package validator

import (
	"fmt"
	"kwil/pkg/databases"
	anytype "kwil/pkg/types/data_types/any_type"
)

/*
###################################################################################################

	Queries: 700-799

###################################################################################################
*/

func (v *Validator) validateQueries() error {
	if len(v.db.SQLQueries) > databases.MAX_QUERY_COUNT {
		return violation(errorCode701, fmt.Errorf(`database has too many queries: %v > %v`, len(v.db.SQLQueries), databases.MAX_QUERY_COUNT))
	}

	queryNames := make(map[string]struct{})
	for _, q := range v.db.SQLQueries {
		if _, ok := queryNames[q.Name]; ok {
			return violation(errorCode700, fmt.Errorf(`duplicate query name "%s"`, q.Name))
		}
		queryNames[q.Name] = struct{}{}

		err := v.ValidateQuery(q)
		if err != nil {
			return fmt.Errorf(`error on query %v: %w`, q.Name, err)
		}
	}

	return nil
}

/*
###################################################################################################

	Query: 800-899

###################################################################################################
*/

func (v *Validator) ValidateQuery(q *databases.SQLQuery[anytype.KwilAny]) error {
	if err := CheckName(q.Name, databases.MAX_QUERY_NAME_LENGTH); err != nil {
		return violation(errorCode800, err)
	}

	if isReservedWord(q.Name) {
		return violation(errorCode808, fmt.Errorf(`query name "%s" is a reserved word`, q.Name))
	}

	if !q.Type.IsValid() {
		return violation(errorCode801, fmt.Errorf(`invalid query type "%d"`, q.Type))
	}

	table := v.db.GetTable(q.Table)
	if table == nil {
		return violation(errorCode802, fmt.Errorf(`table "%s" does not exist`, q.Table))
	}

	if q.Type == databases.INSERT || q.Type == databases.UPDATE {
		if len(q.Params) == 0 {
			return violation(errorCode803, fmt.Errorf(`query "%s" must have at least one parameter`, q.Name))
		}
	}

	if q.Type == databases.UPDATE || q.Type == databases.DELETE {
		if len(q.Where) == 0 {
			return violation(errorCode804, fmt.Errorf(`query "%s" must have at least one where clause`, q.Name))
		}
	}

	if q.Type == databases.INSERT && len(q.Where) > 0 {
		return violation(errorCode805, fmt.Errorf(`insert query "%s" cannot have where clauses`, q.Name))
	}

	if q.Type == databases.DELETE && len(q.Params) > 0 {
		return violation(errorCode806, fmt.Errorf(`delete query "%s" cannot have parameters`, q.Name))
	}

	if !allNotNullHaveParamForInsert(q, table) {
		return violation(errorCode807, fmt.Errorf(`insert query "%s" must have a parameter for every non-static column`, q.Name))
	}

	return v.validateInputs(q.Params, q.Where, table)
}

// allNotNullHaveParamForInsert checks that all non-static columns have a parameter for insert queries
func allNotNullHaveParamForInsert(q *databases.SQLQuery[anytype.KwilAny], table *databases.Table[anytype.KwilAny]) bool {
	if q.Type != databases.INSERT {
		return true
	}

	for _, c := range table.Columns {
		if c.GetAttribute(databases.NOT_NULL) == nil {
			continue
		}

		found := false
		for _, p := range q.Params {
			if p.Column == c.Name {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
