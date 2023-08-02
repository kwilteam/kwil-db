package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/types/validation"
)

// runCleans runs a series of clean functions and returns the first error.
func runCleans(errs ...error) error {
	return errors.Join(errs...)
}

func cleanIdent(ident *string) error {
	if ident == nil {
		return fmt.Errorf("identifier cannot be nil")
	}

	err := validation.ValidateIdentifier(*ident)
	if err != nil {
		return err
	}

	*ident = strings.ToLower(*ident)

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

func cleanScalar(scalar *any) error {
	if scalar == nil {
		return nil
	}

	kind := reflect.TypeOf(*scalar).Kind()

	switch kind {
	case reflect.String, reflect.Int, reflect.Float64, reflect.Bool, reflect.Int64, reflect.Float32, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return nil
	default:
		return fmt.Errorf("invalid scalar type: %s", kind.String())
	}
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

	return nil
}
