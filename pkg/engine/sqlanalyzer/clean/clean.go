/*
Package clean cleans SQL queries.

This includes making identifiers lower case.

The walker in this package implements all the tree.Walker methods, even if it
doesn't do anything. This is to ensure that if we need to add more cleaning / validation
rules, we know that we've covered all the nodes.

For example, EnterDelete does nothing, but if we later set a limit on the amount of
CTEs allowed, then we would add it there.
*/
package clean

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer/utils"
)

// checks that the string starts with a letter
var startsWithLetter = regexp.MustCompile(`^[a-zA-Z]`)

// checks that the string only contains alphanumeric characters and underscores
var onlyAlnumUnderscore = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// cleanIdentifier checks that the identifier is a valid identifier and returns
// it in lower case
func cleanIdentifier(identifier string) (string, error) {
	res := strings.ToLower(identifier)

	if !startsWithLetter.MatchString(res) {
		return "", wrapErr(ErrInvalidIdentifier, fmt.Errorf(`identifier must start with a letter, received: "%s"`, identifier))
	}

	if !onlyAlnumUnderscore.MatchString(res) {
		return "", wrapErr(ErrInvalidIdentifier, fmt.Errorf(`identifier must only contain alphanumeric characters or underscores, received: "%s"`, identifier))
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

// checkLiteral checks that the literal is a valid literal.
// It either must be guarded with single quotes, or it must be a number.
func checkLiteral(literal string) error {
	_, err := utils.IsLiteral(literal)
	return wrapErr(ErrInvalidLiteral, err)
}

// checkBindParameter checks that the bind parameter is a valid bind parameter.
// It must start with either a $ or @.
func checkBindParameter(bindParameter string) error {
	if !strings.HasPrefix(bindParameter, "$") && !strings.HasPrefix(bindParameter, "@") {
		return wrapErr(ErrInvalidBindParameter, errors.New("bind parameter must start with $ or @"))
	}

	return nil
}
