package models

import (
	"fmt"
	types "kwil/x/sqlx"
	ddlb "kwil/x/sqlx/ddl_builder"
)

type Table struct {
	Name    string    `json:"name"`
	Columns []*Column `json:"columns"`
}

// Validate checks if the table is valid.
// The DB is passed to make this fullfill the Definition interface.
func (t *Table) Validate(db *Database) error {
	// check if table name is valid
	err := CheckName(t.Name, types.TABLE)
	if err != nil {
		return err
	}

	// check column name uniqueness
	columns := make(map[string]struct{})
	for _, col := range t.Columns {
		if _, ok := columns[col.Name]; ok {
			return fmt.Errorf(`duplicate column name "%s"`, col.Name)
		}
		columns[col.Name] = struct{}{}

		// check if column is valid
		err := col.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Table) GetColumn(name string) *Column {
	for _, col := range t.Columns {
		if col.Name == name {
			return col
		}
	}
	return nil
}

func (t *Table) GenerateDDL(schemaName string) ([]string, error) {
	tbl := ddlb.NewTableBuilder().Schema(schemaName).Name(t.Name)
	for _, col := range t.Columns {
		cb := ddlb.NewColumnBuilder()

		// convert column type to Postgres type
		pgtype, err := types.Conversion.KwilStringToPgType(col.Type)
		if err != nil {
			return nil, err
		}

		column := cb.Name(col.Name).Type(pgtype)

		for _, attribute := range col.Attributes {
			column.WithAttribute(attribute.Type, attribute.Value)
		}

		tbl.AddColumn(column)
	}

	return tbl.Build()
}
