package validator

import (
	"fmt"
	"kwil/pkg/databases"
	anytype "kwil/pkg/types/data_types/any_type"
)

// validation errorCodes are annotated with errorCode number. i.e. "errorCode 1"

// a validator validates a database
type Validator struct {
	db *databases.Database[anytype.KwilAny]
}

func New(db *databases.Database[anytype.KwilAny]) *Validator {
	return &Validator{db: db}
}

// Validate validates a database
func (v *Validator) Validate(db *databases.Database[anytype.KwilAny]) error {
	v.db = db

	// errorCode 0
	err := CheckName(db.Name, databases.MAX_DB_NAME_LENGTH)
	if err != nil {
		return violation(errorCode0, err)
	}

	// errorCode 1
	err = CheckAddress(db.Owner)
	if err != nil {
		return violation(errorCode1, err)
	}

	// errorCodes 100-699
	err = v.validateTables()
	if err != nil {
		return fmt.Errorf(`invalid tables: %w`, err)
	}

	// errorCodes 700-1099
	err = v.validateQueries()
	if err != nil {
		return fmt.Errorf(`invalid queries: %w`, err)
	}

	// errorCodes 1100-1299
	err = v.validateIndexes()
	if err != nil {
		return fmt.Errorf(`invalid indexes: %w`, err)
	}

	// errorCodes 1300-1499
	err = v.validateRoles()
	if err != nil {
		return fmt.Errorf(`invalid roles: %w`, err)
	}

	return nil
}
