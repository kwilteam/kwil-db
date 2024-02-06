package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/types/validation"
)

// runCleans runs a series of clean functions and returns the first error.
func runCleans(errs ...error) error {
	return errors.Join(errs...)
}

func cleanIdent(ident *string) error {
	err := cleanString(ident)
	if err != nil {
		return err
	}

	err = validation.ValidateIdentifier(*ident)
	if err != nil {
		return err
	}

	return nil
}

func cleanDBID(dbid *string) error {
	err := cleanString(dbid)
	if err != nil {
		return err
	}

	err = validation.ValidateDBID(*dbid)
	if err != nil {
		return err
	}

	return nil
}

// cleanString cleans a string by trimming whitespace and making it lowercase.
// It returns an error if the string is nil.
func cleanString(str *string) error {
	if str == nil {
		return fmt.Errorf("string cannot be nil")
	}

	*str = strings.TrimSpace(*str)
	*str = strings.ToLower(*str)

	return nil
}

func cleanIdents(idents *[]string) error {
	if idents == nil {
		return fmt.Errorf("identifiers cannot be nil")
	}

	for i := range *idents {
		err := cleanIdent(&(*idents)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanActionParameters(inputs *[]string) error {
	if inputs == nil {
		return nil
	}

	for i := range *inputs {
		err := cleanActionParameter(&(*inputs)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanActionParameter applies only to the unparsed instructions/statements.
func cleanActionParameter(input *string) error {
	if len(*input) == 0 {
		return fmt.Errorf("action parameter cannot be empty")
	}

	if len(*input) > validation.MAX_IDENT_NAME_LENGTH {
		return fmt.Errorf("action parameter cannot be longer than %d characters", validation.MAX_IDENT_NAME_LENGTH)
	}

	if !strings.HasPrefix(*input, "$") {
		return fmt.Errorf("action parameter must start with $")
	}

	*input = strings.ToLower(*input)

	return nil
}
