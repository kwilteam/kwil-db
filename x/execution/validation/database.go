package validation

import (
	"fmt"
	"kwil/x/execution"
	"kwil/x/types/databases"
)

func ValidateDatabase(db *databases.Database) error {
	// check if database name is valid
	err := CheckName(db.Name, execution.MAX_DB_NAME_LENGTH)
	if err != nil {
		return fmt.Errorf(`invalid database name: %w`, err)
	}

	// check owner name (this is sort of redundant, but it's here for consistency)
	err = CheckAddress(db.Owner)
	if err != nil {
		return fmt.Errorf(`invalid owner name: %w`, err)
	}

	// validate tables
	err = validateTables(db)
	if err != nil {
		return fmt.Errorf(`error on tables: %w`, err)
	}

	// validate roles
	err = validateRoles(db)
	if err != nil {
		return fmt.Errorf(`error on roles: %w`, err)
	}

	// validate SQLqueries
	err = validateSQLQueries(db)
	if err != nil {
		return fmt.Errorf(`error on SQL queries: %w`, err)
	}

	// validate indexes
	err = validateIndexes(db)
	if err != nil {
		return fmt.Errorf(`error on indexes: %w`, err)
	}

	return nil
}
