package schema

import (
	"fmt"
	"regexp"
)

const (
	MAX_COLUMN_NAME_LENGTH = 32
	MAX_TABLE_NAME_LENGTH  = 32
	MAX_INDEX_NAME_LENGTH  = 32
	MAX_ROLE_NAME_LENGTH   = 32
	MAX_COLUMNS_PER_TABLE  = 50
	MAX_DB_NAME_LENGTH     = 16
	MAX_QUERY_NAME_LENGTH  = 32
)

func (db *Database) Validate() error {
	/*
		// verify must check:
		- DB name is valid

			TABLES:
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

			ROLES:
			- no empty role names
			- role names must begin with a letter or underscore
			- role name is no longer than MAX_ROLE_NAME_LENGTH characters
			- queries must exist
			- queries must be unique
	*/

	// check db name is valid
	if len(db.Name) > MAX_DB_NAME_LENGTH {
		return fmt.Errorf("database name must be less than %d characters", MAX_DB_NAME_LENGTH)
	}
	ok, err := CheckValidName(db.Name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("database name must begin with a letter and contain only letters, numbers, and underscores")
	}

	// verify table names
	for name, table := range db.Tables {
		ok, err := CheckValidName(name)
		if err != nil {
			return fmt.Errorf("error checking allowed characters: %w", err)
		}
		if !ok {
			return fmt.Errorf("table name %s must begin with a letter", name)
		}
		if len(name) > MAX_TABLE_NAME_LENGTH {
			return fmt.Errorf("table name %s is too long (max length is %d)", name, MAX_TABLE_NAME_LENGTH)
		}
		err = table.Validate()
		if err != nil {
			return fmt.Errorf("error on table %s: %w", name, err)
		}
	}

	// verify indexes
	indexTracker := make(map[string]string)
	for name, index := range db.Indexes {
		ok, err := CheckValidName(name)
		if err != nil {
			return fmt.Errorf("error checking allowed characters: %w", err)
		}
		if !ok {
			return fmt.Errorf("index name %s must begin with a letter", name)
		}
		if len(name) > MAX_INDEX_NAME_LENGTH {
			return fmt.Errorf("index name %s is too long (max length is %d)", name, MAX_INDEX_NAME_LENGTH)
		}

		// check for duplicate indexes
		key := fmt.Sprintf("%s.%s.%s", index.Table, index.Column, index.Using)
		if indexTracker[key] != "" {
			return fmt.Errorf("duplicate index %s and %s both index %s.%s", indexTracker[key], name, index.Table, index.Column)
		}
		indexTracker[key] = name
		err = index.Validate(db)
		if err != nil {
			return fmt.Errorf("error on index %s: %w", name, err)
		}
	}

	// verify roles
	for name, role := range db.Roles {
		ok, err := CheckValidName(name)
		if err != nil {
			return fmt.Errorf("error checking allowed characters: %w", err)
		}
		if !ok {
			return fmt.Errorf("role name %s must begin with a letter", name)
		}
		if len(name) > MAX_ROLE_NAME_LENGTH {
			return fmt.Errorf("role name %s is too long (max length is %d)", name, MAX_ROLE_NAME_LENGTH)
		}
		err = role.Validate(db)
		if err != nil {
			return fmt.Errorf("error on role %s: %w", name, err)
		}
	}

	// validate that the default role exists
	if db.DefaultRole == "" {
		return fmt.Errorf("default role must be set")
	}
	rs := db.ListRoles()
	found := false
	for _, r := range rs {
		if r == db.DefaultRole {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("default role %s does not exist", db.DefaultRole)
	}

	return nil
}

func (t *Table) Validate() error {
	/*
			verify must check:
		- no empty column names
		- column names must begin with a letter or underscore
		- column name is no longer than MAX_COLUMN_NAME_LENGTH characters
		- Must be <= MAX_COLUMNS_PER_TABLE columns
		- Validate attribute is valid
	*/
	if len(t.Columns) > MAX_COLUMNS_PER_TABLE {
		return fmt.Errorf("table can have at most %d columns", MAX_COLUMNS_PER_TABLE)
	}
	for name, column := range t.Columns {
		ok, err := CheckValidName(name)
		if err != nil {
			return fmt.Errorf("error checking allowed characters: %w", err)
		}
		if !ok {
			return fmt.Errorf("column name %s must begin with a letter", name)
		}
		if len(name) > MAX_COLUMN_NAME_LENGTH {
			return fmt.Errorf("column name %s is too long", name)
		}

		err = column.Validate()
		if err != nil {
			return fmt.Errorf("error on column %s: %w", name, err)
		}
	}
	return nil
}

func (c *KuniformColumn) Validate() error {
	/*
		verify must check:
		- Validate type is valid
		- Validate attribute is valid
	*/

	// see if there is an associated pg type
	if !c.Type.Valid() {
		return fmt.Errorf("invalid type %s", c.Type)
	}

	// validate attributes
	atts, err := c.GetAttributes()
	if err != nil {
		return fmt.Errorf("error getting attributes: %w", err)
	}

	for _, att := range atts {
		if !att.Valid() {
			return fmt.Errorf("invalid attribute %s", att)
		}
	}
	return nil
}

func (i *Index) Validate(db *Database) error {
	/*
		- table must exist
		- column must exist
		- using must be valid
	*/

	// check that table exists
	if _, ok := db.Tables[i.Table]; !ok {
		return fmt.Errorf("table %s does not exist", i.Table)
	}

	// check that column exists
	if _, ok := db.Tables[i.Table].Columns[i.Column]; !ok {
		return fmt.Errorf("column %s does not exist in table %s", i.Column, i.Table)
	}

	// check that using is valid
	if !i.Using.Valid() {
		return fmt.Errorf("unknown index type: %s", i.Using)
	}
	return nil
}

func (r *Role) Validate(db *Database) error {

	// worth noting that this is called in a loop of roles, so this is very inefficient
	// this only gets run at deployment time, so I think we are ok?
	exqs := db.ListQueries()
	for _, q := range r.Queries {
		found := false
		for _, eq := range exqs {
			if eq == q {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("query %s does not exist", q)
		}
	}
	return nil
}

func checkAllowedCharacters(s string) (bool, error) {
	// regex for postgres identifiers
	reg, err := regexp.Compile(`^([[:alpha:]_][[:alnum:]_]*|("[^"]*)+)$`)
	if err != nil {
		return false, err
	}
	return reg.MatchString(s), nil
}

func CheckValidName(s string) (bool, error) {
	if len(s) == 0 {
		return false, nil
	}
	// check if s[0] is an underscore.  The regex will catch all other cases
	if s[0] == '_' {
		return false, nil
	}
	return checkAllowedCharacters(s)
}
