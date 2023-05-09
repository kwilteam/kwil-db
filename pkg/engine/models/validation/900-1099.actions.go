package validation

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"regexp"
)

func validateActions(actions []*models.Action) error {
	names := make(map[string]struct{})
	for _, action := range actions {
		if _, ok := names[action.Name]; ok {
			return violation(errorCode900, fmt.Errorf(`duplicate action name "%s"`, action.Name))
		}
		names[action.Name] = struct{}{}

		err := validateAction(action)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateAction(action *models.Action) error {
	inputs := make(map[string]struct{})
	for _, input := range action.Inputs {
		if _, ok := inputs[input]; ok {
			return violation(errorCode1001, fmt.Errorf(`duplicate input name "%s"`, input))
		}
		inputs[input] = struct{}{}

		err := validateInputName(input)
		if err != nil {
			return err
		}
	}

	if len(action.Statements) == 0 {
		return violation(errorCode1003, fmt.Errorf(`action %q has no statements`, action.Name))
	}

	return nil
}

var inputNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func validateInputName(name string) error {
	// check for dollar sign
	if name[0] != '$' {
		return violation(errorCode1000, fmt.Errorf(`input name %q does not start with a dollar sign`, name))
	}

	// check for invalid characters
	if !inputNameRegexp.MatchString(name[1:]) {
		return violation(errorCode1002, fmt.Errorf(`input name %q contains invalid characters`, name))
	}

	return nil
}
