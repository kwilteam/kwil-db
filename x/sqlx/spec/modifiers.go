package types

import "fmt"

type ModifierType int

// Modifiers
const (
	INVALID_MODIFIER ModifierType = iota
	NO_MODIFIER
	CALLER
)

func (m *ModifierType) String() string {
	switch *m {
	case CALLER:
		return "caller"
	}
	return "unknown"
}

// ConvertModifier converts a string to a ModifierType
func (c *conversion) ConvertModifier(s string) (ModifierType, error) {
	switch s {
	case "caller":
		return CALLER, nil
	case "":
		return NO_MODIFIER, nil
	}
	return INVALID_MODIFIER, fmt.Errorf("unknown modifier type")
}
