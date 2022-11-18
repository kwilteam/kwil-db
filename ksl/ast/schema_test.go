package ast_test

import (
	"ksl/ast"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchema(t *testing.T) {
	data := `
	model A {
		id int  @id
		fk int
		b  B?   @ref(fields: [fk], references: [id])
	  }

	  model B {
		id int @id
		a  A?
	  }
`
	s := ast.Parse([]byte(data), "test.kwil")
	require.Empty(t, s.Diagnostics)
}
