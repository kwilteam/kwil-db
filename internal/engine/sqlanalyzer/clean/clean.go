/*
Package clean cleans SQL queries.

This includes making identifiers lower case.

The walker in this package implements all the tree.Walker methods, even if it
doesn't do anything. This is to ensure that if we need to add more cleaning / validation
rules, we know that we've covered all the nodes.

For example, EnterDeleteStmt does nothing, but if we later set a limit on the amount of
CTEs allowed, then we would add it there.
*/
package clean

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kwilteam/kwil-db/common/validation"
)

// checks that the string only contains alphanumeric characters and underscores
var identifierRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// cleanIdentifier checks that the identifier is a valid identifier and returns
// it in lower case
func cleanIdentifier(identifier string) (string, error) {
	res := strings.ToLower(identifier)

	if !identifierRegexp.MatchString(res) {
		return "", wrapErr(ErrInvalidIdentifier, fmt.Errorf(`identifier must start with letter and only contain alphanumeric characters or underscores, received: "%s"`, identifier))
	}

	if validation.IsKeyword(res) {
		return "", wrapErr(ErrInvalidIdentifier, fmt.Errorf(`identifier must not be a keyword, received: "%s"`, identifier))
	}

	return res, nil
}

// cleanIdentifiers checks several identifiers and returns them in lower case
func cleanIdentifiers(identifiers []string) ([]string, error) {
	res := make([]string, len(identifiers))

	for i, identifier := range identifiers {
		var err error
		res[i], err = cleanIdentifier(identifier)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
