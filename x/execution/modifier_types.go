package execution

type ModifierType int

// Modifiers
const (
	NO_MODIFIER           ModifierType = 0
	INVALID_MODIFIER_TYPE ModifierType = iota + 99
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
	if m == NO_MODIFIER {
		return true
	}
	return m > INVALID_MODIFIER_TYPE && m < END_MODIFIER_TYPE
}
