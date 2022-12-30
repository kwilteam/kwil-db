package execution

import "fmt"

type modifiers struct{}

var Modifiers = &modifiers{}

// ConvertModifier converts a string to a ModifierType
func (c *modifiers) ConvertModifier(s string) (ModifierType, error) {
	switch s {
	case "caller":
		return CALLER, nil
	case "":
		return NO_MODIFIER, nil
	}
	return INVALID_MODIFIER, fmt.Errorf("unknown modifier type")
}
