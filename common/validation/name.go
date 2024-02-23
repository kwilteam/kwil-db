package validation

import (
	"fmt"
	"regexp"
)

var validNameRegex = regexp.MustCompile(`^[a-z]\w*$`)

func ValidateIdentifier(name string) error {
	if len(name) > MAX_IDENT_NAME_LENGTH {
		return fmt.Errorf("name too long: %s", name)
	}
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	ok := validNameRegex.MatchString(name)
	if !ok {
		return fmt.Errorf("name must start with letter, only contain letters, numbers, and underscores, and be lowercase.  received: %s", name)
	}

	if IsKeyword(name) {
		return fmt.Errorf("name cannot be a reserved word: %s", name)
	}

	return nil
}

func ValidateDBID(dbid string) error {
	if len(dbid) != 57 {
		return fmt.Errorf("dbid too long: %s", dbid)
	}

	ok := validNameRegex.MatchString(dbid)
	if !ok {
		return fmt.Errorf("dbid must start with letter, only contain letters, numbers, and underscores, and be lowercase.  received: %s", dbid)
	}

	if IsKeyword(dbid) {
		return fmt.Errorf("dbid cannot be a reserved word: %s", dbid)
	}

	return nil
}

func CheckAddress(address string) error {
	if len(address) == 0 {
		return fmt.Errorf("address cannot be empty")
	}

	if len(address) > MAX_OWNER_NAME_LENGTH {
		return fmt.Errorf("address must be less than or equal to 44 characters")
	}
	return nil
}
