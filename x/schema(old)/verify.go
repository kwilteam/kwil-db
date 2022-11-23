package schema

import (
	"errors"
	"fmt"
)

const (
	MAX_COLUMN_NAME_LENGTH = 32
	MAX_TABLE_NAME_LENGTH  = 32
	MAX_INDEX_NAME_LENGTH  = 32
	MAX_ROLE_NAME_LENGTH   = 32
	MAX_COLUMNS_PER_TABLE  = 50
)

// Verify verifies all aspects of a database besides the database's name.
func Verify(db *Database[KwilType, KwilConstraint, KwilIndex]) error {
	/*
		verify must check:
		- no empty table names
		- table names must begin with a letter or underscore
		- table name is no longer than MAX_TABLE_NAME_LENGTH characters
		- must verify all tables (using verifyTable function)

		INDEXES:
		- no empty index names
		- index names must begin with a letter or underscore
		- index name is no longer than MAX_INDEX_NAME_LENGTH characters
		- columns can have multiple indexes, but they must be different types
		- must verify all indexes (using verifyIndex function)
	*/

	// check that table names are valid
	for name, table := range db.Tables {
		if len(name) == 0 {
			return fmt.Errorf("empty table name")
		}
		if !startsWithAllowedCharacter(name) {
			return fmt.Errorf("table name %s must begin with a letter or underscore", name)
		}
		if len(name) > MAX_TABLE_NAME_LENGTH {
			return fmt.Errorf("table name %s is too long (max length is %d)", name, MAX_TABLE_NAME_LENGTH)
		}
		err := verifyTable(&table)
		if err != nil {
			return fmt.Errorf("error on table %s: %w", name, err)
		}
	}

	// indexTracker is a map that combines the table name, column name, and index type into a single string
	// and maps it to the index name. This is used to check for duplicate indexes.
	indexTracker := make(map[string]string)
	// check that index names are valid
	for name, index := range db.Indexes {
		if len(name) == 0 {
			return fmt.Errorf("empty index name")
		}
		if !startsWithAllowedCharacter(name) {
			return fmt.Errorf("index name %s must begin with a letter or underscore", name)
		}
		if len(name) > MAX_INDEX_NAME_LENGTH {
			return fmt.Errorf("index name %s is too long (max length is %d)", name, MAX_INDEX_NAME_LENGTH)
		}
		err := verifyIndex(&index, db)
		if err != nil {
			return fmt.Errorf("error on index %s: %w", name, err)
		}

		// check for duplicate indexes
		key := fmt.Sprintf("%s.%s.%s", index.Table, index.Column, index.Using)
		if indexTracker[key] != "" {
			return fmt.Errorf("duplicate index %s and %s both index %s.%s", indexTracker[key], name, index.Table, index.Column)
		}
		indexTracker[key] = name
	}
	return nil
}

// verifyTable verifies all aspects of a table besides the table's name.w
func verifyTable(t *Table[KwilType, KwilConstraint]) error {
	/* verify table must check:
	- no empty column names
	- column names must begin with a letter or underscore
	- column name is no longer than MAX_COLUMN_NAME_LENGTH characters
	- Must be <= MAX_COLUMNS_PER_TABLE columns
	- no duplicate constraint types
	*/
	if len(t.Columns) > MAX_COLUMNS_PER_TABLE {
		return fmt.Errorf("table can have at most %d columns", MAX_COLUMNS_PER_TABLE)
	}
	for name, column := range t.Columns {
		if name == "" {
			return errors.New("empty column name")
		}
		if !startsWithAllowedCharacter(name) {
			return fmt.Errorf("column name %s must begin with a letter or underscore", name)
		}
		if len(name) > MAX_COLUMN_NAME_LENGTH {
			return fmt.Errorf("column name %s is too long", name)
		}

		// check for duplicate constraint types
		constraintTypes := make(map[KwilConstraint]bool)
		for _, constraint := range column.Constraints {
			if constraintTypes[constraint] {
				return fmt.Errorf("duplicate constraint type %s", constraint)
			}
			constraintTypes[constraint] = true
		}
	}
	return nil
}

// verifyIndex verifies all aspects of an index besides the index's name.
// I include the db as a paramater since we must check that tables and columns exist.
func verifyIndex(i *Index[KwilIndex], db *Database[KwilType, KwilConstraint, KwilIndex]) error {
	/*
		- table must exist
		- column must exist
	*/

	// check that table exists
	if _, ok := db.Tables[i.Table]; !ok {
		return fmt.Errorf("table %s does not exist", i.Table)
	}

	// check that column exists
	if _, ok := db.Tables[i.Table].Columns[i.Column]; !ok {
		return fmt.Errorf("column %s does not exist in table %s", i.Column, i.Table)
	}

	return nil
}

func startsWithAllowedCharacter(name string) bool {
	return name[0] == '_' || (name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z')
}
