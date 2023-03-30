package scanner_test

import (
	"kwil/pkg/kl/scanner"
	"kwil/pkg/kl/token"
	"testing"
)

type elt struct {
	tok token.Token
	lit string
}

var tokens = []elt{
	// special tokens

	// identifiers
	{token.IDENT, "foo"},
	{token.IDENT, "bar123"},
	{token.IDENT, "foo_bar"},
	{token.IDENT, "$foo"},
	{token.IDENT, "@bar"},
	{token.INTEGER, "123"},
	{token.STRING, `"hello"`},
	{token.STRING, `"hello\\world\n"`},
	{token.STRING, `'hello\\wor\"ld'`},
	{token.STRING, `'hello@world'`},

	// symbols
	{token.LPAREN, "("},
	{token.RPAREN, ")"},
	{token.LBRACE, "{"},
	{token.RBRACE, "}"},
	{token.COMMA, ","},
	{token.SEMICOLON, ";"},
	{token.ASSIGN, "="},
	{token.ADD, "+"},
	{token.SUB, "-"},
	{token.DIV, "/"},
	{token.MUL, "*"},
	{token.MOD, "%"},
	{token.LSS, "<"},
	{token.GTR, ">"},
	{token.LEQ, "<="},
	{token.GEQ, ">="},
	{token.NEQ, "!="},
	{token.NEQ, "<>"},
	{token.NOT, "!"},

	// keywords
	{token.DATABASE, "database"},
	{token.TABLE, "table"},
	{token.ACTION, "action"},
	{token.PUBLIC, "public"},
	{token.PRIVATE, "private"},
	{token.WITH, "with"},
	{token.REPLACE, "replace"},
	{token.INSERT, "insert"},
	{token.SELECT, "select"},
	{token.UPDATE, "update"},
	{token.DELETE, "delete"},
	{token.DROP, "drop"},
	{token.UNIQUE, "unique"},
	{token.INDEX, "index"},
	{token.PRIMARY, "primary"},
	{token.MIN, "min"},
	{token.MAX, "max"},
	{token.MINLEN, "minlen"},
	{token.MAXLEN, "maxlen"},
	//{token.NULL, "null"},
	{token.NOTNULL, "notnull"},
	{token.DEFAULT, "default"},
}

const whitespace = " \t \n "

var source = func() []byte {
	var src []byte
	for _, tok := range tokens {
		src = append(src, tok.lit...)
		src = append(src, whitespace...)
	}
	return src
}()

func TestScanner_Next(t *testing.T) {
	type fields struct {
		src []byte
	}
	tests := []struct {
		name    string
		fields  fields
		wantPos token.Pos
		wantTok token.Token
		wantLit string
	}{
		{
			name: "test support tokens",
			fields: fields{
				src: source,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := token.File{
				Size:  len(tt.fields.src),
				Lines: []int{0},
			}
			s := scanner.New(tt.fields.src, nil, &file)

			pos := 0
			for _, tok := range tokens {
				gotTok, gotLit, gotPos := s.Next()
				if gotTok != tok.tok {
					t.Errorf("Next() gotTok = %v, want %v", gotTok, tok.tok)
				}
				if gotLit != tok.lit {
					t.Errorf("Next() gotLit = %q, want %q", gotLit, tok.lit)
				}

				if gotPos != token.Pos(pos) {
					t.Errorf("Next() gotPos = %v, want %v", gotPos, pos)
				}

				pos += len(tok.lit) + len(whitespace)
			}

			if s.ErrorCount != 0 {
				t.Errorf("got %d errors", s.ErrorCount)
			}
		})
	}

}

func TestScanner_Next_seq(t *testing.T) {
	dbInput := `database test;`
	tableInput := `
table user {
user_id int notnull,
username text,
age int min(18) max(60),
uuid uuid,
gender bool,
email string maxlen(50) minlen(10)
}`
	actionInput := `
action create_user($name, $age) public {
INSERT into user (name, email, wallet) values ($name, "a@b.com", @caller);
}
`
	input := dbInput + tableInput + actionInput

	tests := []struct {
		Type    token.Token
		Literal string
	}{
		{Type: token.DATABASE, Literal: "database"},
		{Type: token.IDENT, Literal: "test"},
		{Type: token.SEMICOLON, Literal: ";"},
		//table
		{Type: token.TABLE, Literal: "table"},
		{Type: token.IDENT, Literal: "user"},
		{Type: token.LBRACE, Literal: "{"},

		{Type: token.IDENT, Literal: "user_id"},
		{Type: token.IDENT, Literal: "int"},
		{Type: token.NOTNULL, Literal: "notnull"},
		{Type: token.COMMA, Literal: ","},

		{Type: token.IDENT, Literal: "username"},
		{Type: token.IDENT, Literal: "text"},
		{Type: token.COMMA, Literal: ","},

		{Type: token.IDENT, Literal: "age"},
		{Type: token.IDENT, Literal: "int"},
		{Type: token.MIN, Literal: "min"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.INTEGER, Literal: "18"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.MAX, Literal: "max"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.INTEGER, Literal: "60"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.COMMA, Literal: ","},

		{Type: token.IDENT, Literal: "uuid"},
		{Type: token.IDENT, Literal: "uuid"},
		{Type: token.COMMA, Literal: ","},

		{Type: token.IDENT, Literal: "gender"},
		{Type: token.IDENT, Literal: "bool"},
		{Type: token.COMMA, Literal: ","},

		{Type: token.IDENT, Literal: "email"},
		{Type: token.IDENT, Literal: "string"},
		{Type: token.MAXLEN, Literal: "maxlen"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.INTEGER, Literal: "50"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.MINLEN, Literal: "minlen"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.INTEGER, Literal: "10"},
		{Type: token.RPAREN, Literal: ")"},

		{Type: token.RBRACE, Literal: "}"},

		//action
		{Type: token.ACTION, Literal: "action"},
		{Type: token.IDENT, Literal: "create_user"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.IDENT, Literal: "$name"},
		{Type: token.COMMA, Literal: ","},
		{Type: token.IDENT, Literal: "$age"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.PUBLIC, Literal: "public"},
		{Type: token.LBRACE, Literal: "{"},
		//
		{Type: token.INSERT, Literal: "INSERT"},
		{Type: token.IDENT, Literal: "into"},
		{Type: token.IDENT, Literal: "user"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.IDENT, Literal: "name"},
		{Type: token.COMMA, Literal: ","},
		{Type: token.IDENT, Literal: "email"},
		{Type: token.COMMA, Literal: ","},
		{Type: token.IDENT, Literal: "wallet"},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.IDENT, Literal: "values"},
		{Type: token.LPAREN, Literal: "("},
		{Type: token.IDENT, Literal: `$name`},
		{Type: token.COMMA, Literal: ","},
		{Type: token.STRING, Literal: `"a@b.com"`},
		{Type: token.COMMA, Literal: ","},
		{Type: token.IDENT, Literal: `@caller`},
		{Type: token.RPAREN, Literal: ")"},
		{Type: token.SEMICOLON, Literal: ";"},

		{Type: token.RBRACE, Literal: "}"},
	}

	file := token.File{
		Size:  len(input),
		Lines: []int{0},
	}
	s := scanner.New([]byte(input), nil, &file)

	for _, tt := range tests {
		tok, lit, _ := s.Next()
		if tok != tt.Type {
			t.Errorf("Next() type wrong, Tok = %q(%v), want %q(%v)", tok, lit, tt.Type, tt.Literal)
		}
		if lit != tt.Literal {
			t.Errorf("Next() literal wrong, Lit = %v, want %v", lit, tt.Literal)
		}
	}

	if s.ErrorCount != 0 {
		t.Errorf("got %d errors", s.ErrorCount)
	}

	cur := 3
	actPos := len(dbInput) + len(tableInput) + cur
	p := file.Position(token.Pos(actPos))
	if p.Line != 10 {
		t.Errorf("wrong line number, got %d, want %d", p.Line, 0)
	}
	if int(p.Column) != cur {
		t.Errorf("wrong column number, got %d, want %d", p.Column, cur)
	}
}

//func TestScanner_Next_error(t *testing.T) {
//
//}
