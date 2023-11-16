package types

import (
	"fmt"
	"strings"
)

type AttributeType string

const (
	PRIMARY_KEY AttributeType = "PRIMARY_KEY"
	UNIQUE      AttributeType = "UNIQUE"
	NOT_NULL    AttributeType = "NOT_NULL"
	DEFAULT     AttributeType = "DEFAULT"
	MIN         AttributeType = "MIN"
	MAX         AttributeType = "MAX"
	MIN_LENGTH  AttributeType = "MIN_LENGTH"
	MAX_LENGTH  AttributeType = "MAX_LENGTH"
)

func (a AttributeType) String() string {
	return string(a)
}

func (a *AttributeType) IsValid() bool {
	upper := strings.ToUpper(a.String())

	return upper == PRIMARY_KEY.String() ||
		upper == UNIQUE.String() ||
		upper == NOT_NULL.String() ||
		upper == DEFAULT.String() ||
		upper == MIN.String() ||
		upper == MAX.String() ||
		upper == MIN_LENGTH.String() ||
		upper == MAX_LENGTH.String()
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (a *AttributeType) Clean() error {
	if !a.IsValid() {
		return fmt.Errorf("invalid attribute type: %s", a.String())
	}

	*a = AttributeType(strings.ToUpper(a.String()))

	return nil
}
