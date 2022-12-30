package execution

type AttributeType int

// Attributes
const (
	INVALID_ATTRIBUTE AttributeType = iota
	PRIMARY_KEY
	UNIQUE
	NOT_NULL
	DEFAULT
	MIN        // Min allowed value
	MAX        // Max allowed value
	MIN_LENGTH // Min allowed length
	MAX_LENGTH // Max allowed length
	END_ATTRIBUTE
)

func (a *AttributeType) String() string {
	switch *a {
	case PRIMARY_KEY:
		return `primary_key`
	case UNIQUE:
		return `unique`
	case NOT_NULL:
		return `not_null`
	case DEFAULT:
		return `default`
	case MIN:
		return `min`
	case MAX:
		return `max`
	case MIN_LENGTH:
		return `min_length`
	case MAX_LENGTH:
		return `max_length`
	}
	return `unknown`
}

func (a *AttributeType) Int() int {
	return int(*a)
}

func (a *AttributeType) IsValid() bool {
	return *a > INVALID_ATTRIBUTE && *a < END_ATTRIBUTE
}
