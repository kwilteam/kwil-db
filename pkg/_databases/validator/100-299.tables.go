package validator

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

/*
###################################################################################################

	Tables: 100-199

###################################################################################################
*/

// errorCodes 100-199
func (v *Validator) validateTables() error {
	// errorCodes 101 and 102
	if err := v.validateTableCount(); err != nil {
		return fmt.Errorf(`invalid table count: %w`, err)
	}

	// validate tables
	tablesNames := make(map[string]struct{})
	for _, tbl := range v.DB.Tables {
		if _, ok := tablesNames[tbl.Name]; ok {
			return violation(errorCode100, fmt.Errorf(`duplicate table name "%s"`, tbl.Name))
		}
		tablesNames[tbl.Name] = struct{}{}

		// errorCodes 200-299
		err := v.ValidateTable(tbl)
		if err != nil {
			return fmt.Errorf(`error on table %v: %w`, tbl.Name, err)
		}
	}

	return nil
}

// validateTableCount validates the number of tables in the database
// is within the allowed range
func (v *Validator) validateTableCount() error {
	// errorCode 101
	if len(v.DB.Tables) == 0 {
		return violation(errorCode101, fmt.Errorf(`database has 0 tables`))
	}

	// errorCode 102
	if len(v.DB.Tables) > MAX_TABLE_COUNT {
		return violation(errorCode102, fmt.Errorf(`database has too many tables: %v > %v`, len(v.DB.Tables), MAX_TABLE_COUNT))
	}

	return nil
}

/*
###################################################################################################

	Table: 200-299

###################################################################################################
*/

func (v *Validator) ValidateTable(tbl *databases.Table[*spec.KwilAny]) error {
	if err := CheckName(tbl.Name, MAX_TABLE_NAME_LENGTH); err != nil {
		return violation(errorCode200, err)
	}

	if isReservedWord(tbl.Name) {
		return violation(errorCode201, fmt.Errorf(`table name "%s" is a reserved word`, tbl.Name))
	}

	return v.validateColumns(tbl.Columns)
}
