package validation

import (
	"fmt"
	"kwil/pkg/engine/models"
)

/*
###################################################################################################

	Tables: 100-199

###################################################################################################
*/

// errorCodes 100-199
func validateTables(tables []*models.Table) error {
	// errorCodes 101 and 102
	if err := validateTableCount(tables); err != nil {
		return fmt.Errorf(`invalid table count: %w`, err)
	}

	// validate tables
	tablesNames := make(map[string]struct{})
	for _, tbl := range tables {
		if _, ok := tablesNames[tbl.Name]; ok {
			return violation(errorCode100, fmt.Errorf(`duplicate table name "%s"`, tbl.Name))
		}
		tablesNames[tbl.Name] = struct{}{}

		// errorCodes 200-299
		err := ValidateTable(tbl)
		if err != nil {
			return fmt.Errorf(`error on table %v: %w`, tbl.Name, err)
		}
	}

	return nil
}

// validateTableCount validates the number of tables in the database
// is within the allowed range
func validateTableCount(tables []*models.Table) error {
	// errorCode 101
	if len(tables) == 0 {
		return violation(errorCode101, fmt.Errorf(`database has 0 tables`))
	}

	// errorCode 102
	if len(tables) > MAX_TABLE_COUNT {
		return violation(errorCode102, fmt.Errorf(`database has too many tables: %v > %v`, len(tables), MAX_TABLE_COUNT))
	}

	return nil
}

/*
###################################################################################################

	Table: 200-299

###################################################################################################
*/

func ValidateTable(tbl *models.Table) error {
	if err := CheckName(tbl.Name, MAX_TABLE_NAME_LENGTH); err != nil {
		return violation(errorCode200, err)
	}

	if isReservedWord(tbl.Name) {
		return violation(errorCode201, fmt.Errorf(`table name "%s" is a reserved word`, tbl.Name))
	}

	if err := validateIndexes(tbl); err != nil {
		return fmt.Errorf(`error on table %s indexes: %w`, tbl.Name, err)
	}

	return validateColumns(tbl.Columns)
}
