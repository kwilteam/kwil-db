package validation

import (
	"fmt"
	"kwil/pkg/engine/models"
)

/*
###################################################################################################

	Columns: 300-399

###################################################################################################
*/

// validateColumns validates all columns in an array
func validateColumns(columns []*models.Column) error {
	if len(columns) > MAX_COLUMNS_PER_TABLE {
		return violation(errorCode301, fmt.Errorf(`too many columns: %v > %v`, len(columns), MAX_COLUMNS_PER_TABLE))
	}

	columnNames := make(map[string]struct{})
	for _, col := range columns {
		if _, ok := columnNames[col.Name]; ok {
			return violation(errorCode300, fmt.Errorf(`duplicate column name %q`, col.Name))
		}
		columnNames[col.Name] = struct{}{}

		if err := ValidateColumn(col); err != nil {
			return err
		}
	}
	return nil
}

/*
###################################################################################################

	Column: 400-499

###################################################################################################
*/

// ValidateColumn validates a single column
func ValidateColumn(col *models.Column) error {
	if err := CheckName(col.Name, MAX_COLUMN_NAME_LENGTH); err != nil {
		return violation(errorCode400, err)
	}

	if isReservedWord(col.Name) {
		return violation(errorCode401, fmt.Errorf(`column name %q is a reserved word`, col.Name))
	}

	if !col.Type.IsValid() {
		return violation(errorCode402, fmt.Errorf(`invalid column type %q`, col.Type))
	}

	return validateAttributes(col)
}
