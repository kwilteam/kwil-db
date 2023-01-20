package convert

import (
	"fmt"
	"kwil/x/execution"
	datatypes "kwil/x/types/data_types"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
	"strings"
)

type clean struct{}

// clean database lowercases necessary fields and converts the database to a
// kwilAny database.
func (c clean) CleanDatabase(db *databases.Database[[]byte]) (*databases.Database[anytype.KwilAny], error) {
	db.Name = strings.ToLower(db.Name)
	db.Owner = strings.ToLower(db.Owner)
	tables := make([]*databases.Table[anytype.KwilAny], len(db.Tables))
	for i, tbl := range db.Tables {
		var err error
		table, err := c.CleanTable(tbl)
		if err != nil {
			return nil, fmt.Errorf("error cleaning table %s: %w", tbl.Name, err)
		}

		tables[i] = table
	}
	for _, qry := range db.SQLQueries {
		c.CleanSQLQuery(qry)
	}
	for _, role := range db.Roles {
		c.CleanRole(role)
	}
	for _, index := range db.Indexes {
		c.CleanIndex(index)
	}

	return &databases.Database[anytype.KwilAny]{
		Name:       db.Name,
		Owner:      db.Owner,
		Tables:     tables,
		SQLQueries: db.SQLQueries,
		Roles:      db.Roles,
		Indexes:    db.Indexes,
	}, nil
}

// Clean cleans the table.
func (c clean) CleanTable(tbl *databases.Table[[]byte]) (*databases.Table[anytype.KwilAny], error) {
	tbl.Name = strings.ToLower(tbl.Name)
	cols := make([]*databases.Column[anytype.KwilAny], len(tbl.Columns))

	for i, col := range tbl.Columns {
		var err error
		cols[i], err = c.CleanColumn(col)
		if err != nil {
			return nil, fmt.Errorf("error in table %s: %w", tbl.Name, err)
		}
	}

	return &databases.Table[anytype.KwilAny]{
		Name:    tbl.Name,
		Columns: cols,
	}, nil
}

// Clean cleans the column.
func (c clean) CleanColumn(col *databases.Column[[]byte]) (*databases.Column[anytype.KwilAny], error) {
	col.Name = strings.ToLower(col.Name)
	if col.Type > datatypes.END_DATA_TYPE || col.Type < datatypes.INVALID_DATA_TYPE { // this should get caught by validation, but just in case
		col.Type = datatypes.INVALID_DATA_TYPE
	}
	attributes := make([]*databases.Attribute[anytype.KwilAny], len(col.Attributes))
	for i, attr := range col.Attributes {
		if attr.Type > execution.END_ATTRIBUTE_TYPE || attr.Type < execution.INVALID_ATTRIBUTE_TYPE { // this should get caught by validation, but just in case
			attr.Type = execution.INVALID_ATTRIBUTE_TYPE
		}

		// convert the attribute to a kwilAny attribute
		anyAttr, err := anytype.NewFromSerial(attr.Value)
		if err != nil {
			return nil, fmt.Errorf("error converting attribute %s: %w", attr.Type.String(), err)
		}

		attributes[i] = &databases.Attribute[anytype.KwilAny]{
			Type:  attr.Type,
			Value: anyAttr,
		}
	}

	return &databases.Column[anytype.KwilAny]{
		Name:       col.Name,
		Type:       col.Type,
		Attributes: attributes,
	}, nil
}

// Clean cleans the role.
func (c clean) CleanRole(role *databases.Role) {
	role.Name = strings.ToLower(role.Name)
	for i, val := range role.Permissions {
		role.Permissions[i] = strings.ToLower(val)
	}
}

// Clean cleans the SQL query.
func (c clean) CleanSQLQuery(qry *databases.SQLQuery) {
	qry.Name = strings.ToLower(qry.Name)
	if qry.Type > execution.END_QUERY_TYPE || qry.Type < execution.INVALID_QUERY_TYPE { // this should get caught by validation, but just in case
		qry.Type = execution.INVALID_QUERY_TYPE
	}

	qry.Table = strings.ToLower(qry.Table)
	for _, param := range qry.Params {
		c.CleanParam(param)
	}
	for _, where := range qry.Where {
		c.CleanWheres(where)
	}
}

// Clean cleans the param.
func (c clean) CleanParam(param *databases.Parameter) {
	param.Name = strings.ToLower(param.Name)
	param.Column = strings.ToLower(param.Column)
	if param.Modifier > execution.END_MODIFIER_TYPE || param.Modifier < execution.INVALID_MODIFIER_TYPE { // this should get caught by validation, but just in case
		param.Modifier = execution.INVALID_MODIFIER_TYPE
	}
}

// Clean cleans the where predicate.
func (c clean) CleanWheres(where *databases.WhereClause) {
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
func (c clean) CleanIndex(i *databases.Index) {
	i.Name = strings.ToLower(i.Name)
	i.Table = strings.ToLower(i.Table)
	if i.Using > execution.END_INDEX_TYPE || i.Using < execution.INVALID_INDEX_TYPE { // this should get caught by validation, but just in case
		i.Using = execution.INVALID_INDEX_TYPE
	}
	for j, column := range i.Columns {
		i.Columns[j] = strings.ToLower(column)
	}
}
