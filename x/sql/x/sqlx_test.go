package sqlx_test

import (
	"strconv"
	"testing"

	sqlx "kwil/x/sql/x"

	"github.com/stretchr/testify/require"
)

func TestMayWrap(t *testing.T) {
	tests := []struct {
		input   string
		wrapped bool
	}{
		{"", true},
		{"()", false},
		{"('text')", false},
		{"('(')", false},
		{`('(\\')`, false},
		{`('\')(')`, false},
		{`(a) in (b)`, true},
		{`a in (b)`, true},
		{`("\\\\(((('")`, false},
		{`('(')||(')')`, true},
		// Test examples from SQLite.
		{"b || 'mx'", true},
		{"a+1", true},
		{"substr(mx, 2)", true},
		{"(json_extract(mx, '$.a'))", false},
		{"(substr(a, 2) COLLATE NOCASE)", false},
		{"(b+random())", false},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			expect := tt.input
			if tt.wrapped {
				expect = "(" + expect + ")"
			}
			require.Equal(t, expect, sqlx.MayWrap(tt.input))

		})
	}
}

func TestExprLastIndex(t *testing.T) {
	tests := []struct {
		input   string
		wantIdx int
	}{
		{"", -1},
		{"()", 1},
		{"'('", 2},
		{"('(')", 4},
		{"('text')", 7},
		{"floor(mx), y", 7},
		{"f(floor(mx), y)", 13},
		{"f(floor(mx), y, (z))", 18},
		{"f(mx, (mx*2)), y, (z)", 10},
		{"(a || ' ' || b)", 14},
		{"(a || ', ' || b)", 15},
		{"a || ', ' || b, mx", 13},
		{"(a || ', ' || b), mx", 15},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			idx := sqlx.ExprLastIndex(tt.input)
			require.Equal(t, tt.wantIdx, idx)
		})
	}
}
