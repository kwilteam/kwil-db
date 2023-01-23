package validator

import (
	"fmt"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

/*
	Validate All Columns
*/

// validateColumns validates all columns in an array
func (v *Validator) validateColumns(columns []*databases.Column[anytype.KwilAny]) error {
	// validate column count
	err := validateColumnCount(columns)
	if err != nil {
		return fmt.Errorf(`invalid column count: %w`, err)
	}

	columnNames := make(map[string]struct{})
	for _, col := range columns {
		// validate column name is unique
		if _, ok := columnNames[col.Name]; ok {
			return fmt.Errorf(`duplicate column name "%s"`, col.Name)
		}
		columnNames[col.Name] = struct{}{}

		// validate column
		err := v.validateColumn(col)
		if err != nil {
			return fmt.Errorf(`error on column "%v": %w`, col.Name, err)
		}
	}

	return nil
}

// validateColumnCount validates the number of columns in an array
func validateColumnCount(columns []*databases.Column[anytype.KwilAny]) error {
	if len(columns) > databases.MAX_COLUMNS_PER_TABLE {
		return fmt.Errorf(`too many columns: %v > %v`, len(columns), databases.MAX_COLUMNS_PER_TABLE)
	}

	return nil
}

/*
	Validate Column
*/

// validateColumn validates a single column
func (v *Validator) validateColumn(col *databases.Column[anytype.KwilAny]) error {
	// validate column name
	err := v.validateColumnName(col)
	if err != nil {
		return fmt.Errorf(`invalid column name: %w`, err)
	}

	// validate column type
	err = v.validateColumnType(col)
	if err != nil {
		return fmt.Errorf(`invalid column type: %w`, err)
	}

	err = v.validateAttributes(col)
	if err != nil {
		return fmt.Errorf(`invalid column attributes: %w`, err)
	}

	return nil
}

// validateColumnName validates a column name
func (v *Validator) validateColumnName(col *databases.Column[anytype.KwilAny]) error {
	err := CheckName(col.Name, databases.MAX_COLUMN_NAME_LENGTH)
	if err != nil {
		return fmt.Errorf(`invalid column name: %w`, err)
	}

	return nil
}

// validateColumnType validates a column type
func (v *Validator) validateColumnType(col *databases.Column[anytype.KwilAny]) error {
	if !col.Type.IsValid() {
		return fmt.Errorf(`invalid column type: %v`, col.Type)
	}

	return nil
}
