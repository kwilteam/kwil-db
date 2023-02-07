package validator

import (
	"fmt"
	"kwil/pkg/types/data_types/any_type"
	databases2 "kwil/pkg/types/databases"
)

/*
	Validate All Tables
*/

func (v *Validator) validateTables() error {
	// validate table count
	err := v.validateTableCount()
	if err != nil {
		return fmt.Errorf(`invalid table count: %w`, err)
	}

	// validate tables
	tablesNames := make(map[string]struct{})
	for _, tbl := range v.db.Tables {
		// validate table name is unique
		if _, ok := tablesNames[tbl.Name]; ok {
			return fmt.Errorf(`duplicate table name "%s"`, tbl.Name)
		}
		tablesNames[tbl.Name] = struct{}{}

		err := v.validateTable(tbl)
		if err != nil {
			return fmt.Errorf(`error on table %v: %w`, tbl.Name, err)
		}
	}

	return nil
}

// validateTableCount validates the number of tables in the database
// is within the allowed range
func (v *Validator) validateTableCount() error {
	if len(v.db.Tables) > databases2.MAX_TABLE_COUNT {
		return fmt.Errorf(`too many tables: %v > %v`, len(v.db.Tables), databases2.MAX_TABLE_COUNT)
	}

	return nil
}

/*
	Validate Table
*/

func (v *Validator) validateTable(tbl *databases2.Table[anytype.KwilAny]) error {
	// validate table name
	err := v.validateTableName(tbl)
	if err != nil {
		return fmt.Errorf(`invalid table name: %w`, err)
	}

	// validate columns
	err = v.validateColumns(tbl.Columns)
	if err != nil {
		return fmt.Errorf(`invalid columns: %w`, err)
	}

	return nil
}

// validateTableName validates the name of a table
func (v *Validator) validateTableName(tbl *databases2.Table[anytype.KwilAny]) error {
	err := CheckName(tbl.Name, databases2.MAX_TABLE_NAME_LENGTH)
	if err != nil {
		return fmt.Errorf(`invalid table name: %w`, err)
	}

	return nil
}
