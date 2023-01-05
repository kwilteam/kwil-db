package execution

type ModifierType int

// Modifiers
const (
	INVALID_MODIFIER_TYPE ModifierType = iota - 1 // we start at -1 since modifiers are optional
	NO_MODIFIER
	CALLER
	END_MODIFIER_TYPE
)

func (m ModifierType) String() string {
	switch m {
	case CALLER:
		return "caller"
	}
	return "unknown"
}

func (m ModifierType) Int() int {
	return int(m)
}

func (m ModifierType) IsValid() bool {
	return m > INVALID_MODIFIER_TYPE && m < END_MODIFIER_TYPE
}
