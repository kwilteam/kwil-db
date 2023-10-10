package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/types"
)

// IsLiteral detects if the passed string is convertable to a literal.
// It returns the type of the literal, or an error if it is not a literal.
func IsLiteral(literal string) (types.DataType, error) {
	if strings.HasPrefix(literal, "'") && strings.HasSuffix(literal, "'") {
		return types.TEXT, nil
	}

	if strings.EqualFold(literal, "true") || strings.EqualFold(literal, "false") {
		return types.INT, nil
	}

	if strings.EqualFold(literal, "null") {
		return types.NULL, nil
	}

	_, err := strconv.Atoi(literal)
	if err != nil {
		return types.NULL, fmt.Errorf("invalid literal: could not detect literal type: %s", literal)
	}

	return types.INT, nil
}
