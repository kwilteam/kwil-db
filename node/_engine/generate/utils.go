package generate

import (
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/conv"
)

func attributeToSQLString(col *types.Column, attr *types.Attribute) (string, error) {
	switch attr.Type {
	case types.PRIMARY_KEY:
		return "", nil
	case types.DEFAULT:
		return "DEFAULT " + attr.Value, nil
	case types.NOT_NULL:
		return "NOT NULL", nil
	case types.UNIQUE:
		return "UNIQUE", nil
	case types.MIN:
		return "CHECK (" + col.Name + " >= " + attr.Value + ")", nil
	case types.MAX:
		return "CHECK (" + col.Name + " <= " + attr.Value + ")", nil
	case types.MIN_LENGTH:

		// for max_len and min_len, we want to check that the value is an int.
		// For regular max and min, it can be a decimal or uint256.

		_, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		fn := "LENGTH"
		if col.Type.Equals(types.BlobType) {
			fn = "OCTET_LENGTH"
		}

		return "CHECK (" + fn + "(" + col.Name + ") >= " + attr.Value + ")", nil
	case types.MAX_LENGTH:
		_, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		fn := "LENGTH"
		if col.Type.Equals(types.BlobType) {
			fn = "OCTET_LENGTH"
		}

		return "CHECK (" + fn + "(" + col.Name + ") <= " + attr.Value + ")", nil
	default:
		return "", nil
	}
}

const delimiter = "$kwil_reserved_delim$"

// containsDisallowedDelimiter checks if the string contains the delimiter
func containsDisallowedDelimiter(s string) bool {
	return strings.Contains(s, delimiter)
}
