package sqlparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractIdentifierValue(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{"empty", "", ""},
		{"asymmetric quotes with double quotes", `"a`, `"a`},
		{"asymmetric quotes with bracket quote", `[a`, `[a`},
		{"asymmetric quotes with back tick quote", "`a", "`a"},
		{"double quotes", `"a"`, `a`},
		{"bracket quotes", `[a]`, `a`},
		{"back tick quotes", "`a`", `a`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, extractSQLName(tt.args), "extractSQLName(%v)", tt.args)
		})
	}
}
