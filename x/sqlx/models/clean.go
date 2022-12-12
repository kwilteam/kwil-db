package models

import "strings"

// Clean cleans the database.
// Currently, that just entails lowercasing all the strings (besides default values), but
// in the future, it could do more.
func (db *Database) Clean() {
	db.Name = toLower(db.Name)
	db.Owner = toLower(db.Owner)
	db.DefaultRole = toLower(db.DefaultRole)
	for i := range db.Tables {
		db.Tables[i].Clean()
	}
}

// Clean cleans the table.
func (tbl *Table) Clean() {
	tbl.Name = toLower(tbl.Name)
	for i := range tbl.Columns {
		tbl.Columns[i].Clean()
	}
}

// Clean cleans the column.
func (col *Column) Clean() {
	col.Name = toLower(col.Name)
	col.Type = toLower(col.Type)
	for i := range col.Attributes {
		col.Attributes[i].Clean()
	}
}

// Clean cleans the attribute.
func (attr *Attribute) Clean() {
	attr.Type = toLower(attr.Type)
}

// Clean cleans the role.
func (role *Role) Clean() {
	role.Name = toLower(role.Name)
	for i, val := range role.Permissions {
		role.Permissions[i] = toLower(val)
	}
}

// Clean cleans the SQL query.
func (qry *SQLQuery) Clean() {
	qry.Name = toLower(qry.Name)
	qry.Type = toLower(qry.Type)
	qry.Table = toLower(qry.Table)
	for i := range qry.Params {
		qry.Params[i].Clean()
	}
	for i := range qry.Where {
		qry.Where[i].Clean()
	}
}

// Clean cleans the param.
func (param *Param) Clean() {
	param.Column = toLower(param.Column)
	param.Modifier = toLower(param.Modifier)
	param.Modifier = toLower(param.Modifier)
}

// Clean cleans the where predicate.
func (where *WhereClause) Clean() {
	where.Column = toLower(where.Column)
	where.Operator = toLower(where.Operator)
	where.Modifier = toLower(where.Modifier)
	where.Operator = toLower(where.Operator)
}

// Clean cleans the index.
func (i *Index) Clean() {
	i.Name = toLower(i.Name)
	i.Table = toLower(i.Table)
	i.Using = toLower(i.Using)
	for j, column := range i.Columns {
		i.Columns[j] = toLower(column)
	}
}

func toLower(s string) string {
	return strings.ToLower(s)
}
