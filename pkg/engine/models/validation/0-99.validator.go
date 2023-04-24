package validation

import (
	"fmt"
	"kwil/pkg/engine/models"
)

// validation errorCodes are annotated with errorCode number. i.e. "errorCode 1"

// Validate validates a database
func Validate(db *models.Dataset) error {

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
	err = validateTables(db.Tables)
	if err != nil {
		return fmt.Errorf(`invalid tables: %w`, err)
	}

	err = validateActions(db.Actions)
	if err != nil {
		return fmt.Errorf(`invalid actions: %w`, err)
	}

	return nil
}
