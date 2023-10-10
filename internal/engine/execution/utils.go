package execution

import "fmt"

// utils:
func getIdent(val any) (string, error) {
	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", val)
	}

	err := isIdent(strVal)
	if err != nil {
		return "", err
	}

	return strVal, nil
}

func isIdent(val string) error {
	if len(val) < 2 {
		return fmt.Errorf("expected variable name, got '%s'", val)
	}

	if val[0] != '$' && val[0] != '@' && val[0] != '!' {
		return fmt.Errorf("expected variable name, got '%s'", val)
	}

	return nil
}
