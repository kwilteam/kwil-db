package sqlwriter_test

import (
	"testing"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// most of this gets tested by the tree package, so I am not too worried about coverage here

func Test_Writer(t *testing.T) {
	stmt := sqlwriter.NewWriter()
	stmt.WriteString("TEST1")
	stmt.WriteString("TEST2")

	if stmt.String() != " TEST1  TEST2 " {
		t.Errorf("expected ' TEST1  TEST2 ', got %s", stmt.String())
	}

	stmt = sqlwriter.NewWriter().WrapParen()
	stmt.Token.Lparen()
	stmt.WriteInt64(1)
	stmt.Token.Rparen()
	str := stmt.String()
	if str != "( (  1  ) )" {
		t.Errorf("expected '( (  1  ) ), got %s", stmt.String())
	}

}
