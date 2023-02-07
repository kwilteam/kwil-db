package validator

import (
	"fmt"
	"kwil/pkg/types/data_types/any_type"
	databases2 "kwil/pkg/types/databases"
)

// a validator validates a database
type Validator struct {
	db *databases2.Database[anytype.KwilAny]
}

func (v *Validator) Validate(db *databases2.Database[anytype.KwilAny]) error {
	v.db = db

	// validate name and owner
	err := CheckName(db.Name, databases2.MAX_DB_NAME_LENGTH)
	if err != nil {
		return fmt.Errorf(`invalid database name: %w`, err)
	}

	err = CheckAddress(db.Owner)
	if err != nil {
		return fmt.Errorf(`invalid owner name: %w`, err)
	}

	err = v.validateTables()
	if err != nil {
		return fmt.Errorf(`invalid tables: %w`, err)
	}

	err = v.validateQueries()
	if err != nil {
		return fmt.Errorf(`invalid queries: %w`, err)
	}

	err = v.validateRoles()
	if err != nil {
		return fmt.Errorf(`invalid roles: %w`, err)
	}

	err = v.validateIndexes()
	if err != nil {
		return fmt.Errorf(`invalid indexes: %w`, err)
	}

	return nil
}
