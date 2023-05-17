package tree_test

import (
	"strings"
	"unicode"
)

// this file contains some utils for testing this package specifically

// removeWhitespace removes all whitespace characters from a string.
func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // skip this rune
		}
		return r
	}, s)
}

// compareIgnoringWhitespace compares two strings while ignoring whitespace characters.
func compareIgnoringWhitespace(a, b string) bool {
	aWithoutWhitespace := removeWhitespace(a)
	bWithoutWhitespace := removeWhitespace(b)

	return aWithoutWhitespace == bWithoutWhitespace
}
