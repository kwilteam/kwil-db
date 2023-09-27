package tree

import (
	"fmt"
	"strings"
)

type CollationType string

const (
	CollationTypeBinary CollationType = "BINARY"
	CollationTypeNoCase CollationType = "NOCASE"
	CollationTypeRTrim  CollationType = "RTRIM"
)

func (c CollationType) String() string {
	return string(c)
}

// Valid checks if the collation type is valid.
// Empty collation types are considered valid.
func (c *CollationType) Valid() error {
	if c.Empty() {
		return nil
	}

	newC := CollationType(strings.ToUpper(string(*c)))

	switch newC {
	case CollationTypeBinary, CollationTypeNoCase, CollationTypeRTrim:
	default:
		return fmt.Errorf("invalid collation type: %s", c)
	}
	*c = newC

	return nil
}

func (c CollationType) Empty() bool {
	return c == ""
}
