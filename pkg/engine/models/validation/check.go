package validation

import (
	"fmt"
	"regexp"
	"strings"
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

func isReservedWord(word string) bool {
	upperWord := strings.ToUpper(word)

	_, ok := reservedWords[upperWord]
	return ok
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
