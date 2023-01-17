package clean

import (
	"fmt"
	"kwil/x/execution"
	"kwil/x/types/databases"
	execTypes "kwil/x/types/execution"
	"strings"
)

// Clean cleans the database.
// Currently, that just entails lowercasing all the strings (besides default values), but
// in the future, it could do more.
func CleanDatabase(db *databases.Database) error {
	db.Name = strings.ToLower(db.Name)
	db.Owner = strings.ToLower(db.Owner)
	for _, tbl := range db.Tables {
		err := CleanTable(tbl)
		if err != nil {
			return err
		}
	}
	for _, qry := range db.SQLQueries {
		CleanSQLQuery(qry)
	}
	for _, role := range db.Roles {
		CleanRole(role)
	}
	for _, index := range db.Indexes {
		CleanIndex(index)
	}

	return nil
}

// Clean cleans the table.
func CleanTable(tbl *databases.Table) error {
	tbl.Name = strings.ToLower(tbl.Name)
	for _, col := range tbl.Columns {
		err := CleanColumn(col)
		if err != nil {
			return fmt.Errorf("error in table %s: %w", tbl.Name, err)
		}
	}

	return nil
}

// Clean cleans the column.
func CleanColumn(col *databases.Column) error {
	col.Name = strings.ToLower(col.Name)
	if col.Type > execution.END_DATA_TYPE || col.Type < execution.INVALID_DATA_TYPE { // this should get caught by validation, but just in case
		col.Type = execution.INVALID_DATA_TYPE
	}
	for i := range col.Attributes {
		CleanAttribute(col.Attributes[i])
		err := AssertAttributeType(col.Attributes[i], col.Type)
		if err != nil {
			return fmt.Errorf("error in column %s: %w", col.Name, err)
		}
	}
	return nil
}

// Clean cleans the attribute.
func CleanAttribute(attr *databases.Attribute) {
	if attr.Type > execution.END_ATTRIBUTE_TYPE || attr.Type < execution.INVALID_ATTRIBUTE_TYPE { // this should get caught by validation, but just in case
		attr.Type = execution.INVALID_ATTRIBUTE_TYPE
	}
}

// Clean cleans the role.
func CleanRole(role *databases.Role) {
	role.Name = strings.ToLower(role.Name)
	for i, val := range role.Permissions {
		role.Permissions[i] = strings.ToLower(val)
	}
}

// Clean cleans the SQL query.
func CleanSQLQuery(qry *databases.SQLQuery) {
	qry.Name = strings.ToLower(qry.Name)
	if qry.Type > execution.END_QUERY_TYPE || qry.Type < execution.INVALID_QUERY_TYPE { // this should get caught by validation, but just in case
		qry.Type = execution.INVALID_QUERY_TYPE
	}

	qry.Table = strings.ToLower(qry.Table)
	for _, param := range qry.Params {
		CleanParam(param)
	}
	for _, where := range qry.Where {
		CleanWheres(where)
	}
}

// Clean cleans the param.
func CleanParam(param *databases.Parameter) {
	param.Name = strings.ToLower(param.Name)
	param.Column = strings.ToLower(param.Column)
	if param.Modifier > execution.END_MODIFIER_TYPE || param.Modifier < execution.INVALID_MODIFIER_TYPE { // this should get caught by validation, but just in case
		param.Modifier = execution.INVALID_MODIFIER_TYPE
	}
}

// Clean cleans the where predicate.
func CleanWheres(where *databases.WhereClause) {
	where.Name = strings.ToLower(where.Name)
	where.Column = strings.ToLower(where.Column)
	if where.Modifier > execution.END_MODIFIER_TYPE || where.Modifier < execution.INVALID_MODIFIER_TYPE { // this should get caught by validation, but just in case
		where.Modifier = execution.INVALID_MODIFIER_TYPE
	}

	if where.Operator > execution.END_COMPARISON_OPERATOR_TYPE || where.Operator < execution.INVALID_COMPARISON_OPERATOR_TYPE { // this should get caught by validation, but just in case
		where.Operator = execution.INVALID_COMPARISON_OPERATOR_TYPE
	}
}

// Clean cleans the index.
func CleanIndex(i *databases.Index) {
	i.Name = strings.ToLower(i.Name)
	i.Table = strings.ToLower(i.Table)
	if i.Using > execution.END_INDEX_TYPE || i.Using < execution.INVALID_INDEX_TYPE { // this should get caught by validation, but just in case
		i.Using = execution.INVALID_INDEX_TYPE
	}
	for j, column := range i.Columns {
		i.Columns[j] = strings.ToLower(column)
	}
}

func CleanExecutionBody(body *execTypes.ExecutionBody) {
	body.Database = strings.ToLower(body.Database)
	body.Query = strings.ToLower(body.Query)
}
