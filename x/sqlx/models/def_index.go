package models

import (
	"fmt"
	types "kwil/x/sqlx"
	ddlb "kwil/x/sqlx/ddl_builder"
)

type Index struct {
	Name    string   `json:"name" yaml:"name"`
	Table   string   `json:"table" yaml:"table"`
	Columns []string `json:"columns" yaml:"columns"`
	Using   string   `json:"using" yaml:"using"`
}

func (i *Index) Validate(db *Database) error {
	// check if index name is valid
	err := CheckName(i.Name, types.INDEX)
	if err != nil {
		return err
	}

	// check if index table is valid
	table := db.GetTable(i.Table)
	if table == nil {
		return fmt.Errorf(`table "%s" does not exist`, i.Table)
	}

	// check if index columns are valid
	columns := make(map[string]struct{})
	for _, col := range i.Columns {
		// check if column is unique
		if _, ok := columns[col]; ok {
			return fmt.Errorf(`duplicate column "%s"`, col)
		}
		columns[col] = struct{}{}

		// check if column exists
		if table.GetColumn(col) == nil {
			return fmt.Errorf(`column "%s" does not exist`, col)
		}
	}

	// check if index using is valid
	_, err = types.Conversion.ConvertIndex(i.Using)
	if err != nil {
		return err
	}

	return nil
}

func (i *Index) GenerateDDL(schemaName string) string {
	ib := ddlb.NewIndexBuilder()
	idx := ib.Name(i.Name).Schema(schemaName).Table(i.Table).Columns(i.Columns...).Using(i.Using)
	return idx.Build()
}
