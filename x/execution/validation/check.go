package validation

import (
	"fmt"
	"regexp"
)

var validNameRegex = regexp.MustCompile(`^[a-z]\w*$`)

func CheckName(name string, maxLen int) error {
	if len(name) > maxLen {
		return fmt.Errorf("name too long: %s", name)
	}
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	ok := validNameRegex.MatchString(name)
	if !ok {
		return fmt.Errorf("name must start with letter, only contain letters, numbers, and underscores, and be lowercase.  recieved: %s", name)
	}
	return nil
}

func CheckAddress(address string) error {
	if len(address) == 0 {
		return fmt.Errorf("address cannot be empty")
	}

	if len(address) > 44 || len(address) < 42 {
		return fmt.Errorf("address must be between 42 and 44 characters")
	}
	return nil
}
