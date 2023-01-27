package validator

import (
	"fmt"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

// a validator validates a database
type Validator struct {
	db *databases.Database[anytype.KwilAny]
}

func (v *Validator) Validate(db *databases.Database[anytype.KwilAny]) error {
	v.db = db

	// validate name and owner
	err := CheckName(db.Name, databases.MAX_DB_NAME_LENGTH)
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
