package validator

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

// validation errorCodes are annotated with errorCode number. i.e. "errorCode 1"

// a validator validates a database
type Validator struct {
	DB *databases.Database[*spec.KwilAny]
}

// Validate validates a database
func (v *Validator) Validate(db *databases.Database[*spec.KwilAny]) error {
	v.DB = db

	// errorCode 0
	err := CheckName(db.Name, MAX_DB_NAME_LENGTH)
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
