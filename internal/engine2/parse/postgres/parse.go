package postgres

import (
	"regexp"
)

type CheckSyntaxFunc func(query string) error

var CheckSyntax CheckSyntaxFunc = doNothing

// doNothing is a placeholder for the CheckSyntaxFunc when cgo is disabled.
func doNothing(_ string) error {
	return nil
}

// CheckSyntaxReplaceDollar replaces all bind parameters($x) with 1 to bypass
// syntax check errors.
// () method doesn't convert bind parameters to $1, $2, etc. so we need to
// replace them manually, just so we can do the syntax check.
func CheckSyntaxReplaceDollar(query string) error {
	// Replace all bind parameters($x) with 1 to bypass syntax check errors
	re := regexp.MustCompile(`\$([a-zA-Z_])+`)
	sql := re.ReplaceAllString(query, "1")
	return CheckSyntax(sql)
}
