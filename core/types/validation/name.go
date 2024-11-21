package validation

import (
	"errors"
	"fmt"
	"regexp"
)

var validNameRegex = regexp.MustCompile(`^[a-z]\w*$`)

func ValidateIdentifier(name string) error {
	if len(name) > MAX_IDENT_NAME_LENGTH {
		return fmt.Errorf("name too long: %s", name)
	}
	if len(name) == 0 {
		return errors.New("name cannot be empty")
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
