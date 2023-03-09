package scanner_test

import (
	"kwil/pkg/kl/scanner"
	"kwil/pkg/kl/token"
	"testing"
)

type elt struct {
	tok token.TokenType
	lit string
}

var tokens = []elt{
	// special tokens

	// identifiers
	{token.IDENT, "foo"},
	{token.IDENT, "bar123"},
	{token.IDENT, "foo_bar"},
	{token.INTEGER, "123"},

	// symbols
	{token.LPAREN, "("},
	{token.RPAREN, ")"},
	{token.LBRACE, "{"},
	{token.RBRACE, "}"},
	{token.COMMA, ","},
	{token.SEMICOLON, ";"},

	// keywords
	{token.DATABASE, "database"},
	{token.TABLE, "table"},
	{token.NULL, "null"},
	{token.NOTNULL, "notnull"},
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
		wantTok token.TokenType
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
			var s scanner.Scanner
			s.Init(tt.fields.src)

			for _, tok := range tokens {
				gotLit := s.Next()
				//if gotPos != tt.wantPos {
				//	t.Errorf("Next() gotPos = %v, want %v", gotPos, tt.wantPos)
				//}
				if gotLit.Type != tok.tok {
					t.Errorf("Next() gotTok = %v, want %v", gotLit.Type, tok.tok)
				}
				if gotLit.Literal != tok.lit {
					t.Errorf("Next() gotLit = %v, want %v", gotLit, tok.lit)
				}
			}

			if s.ErrorCount != 0 {
				t.Errorf("got %d errors", s.ErrorCount)
			}
		})
	}

}

//
//func TestScanner_Next2(t *testing.T) {
//
//	input := `table user{user_id int notnull,username string null,age int min(18),gender bool}`
//
//	tests := []token.Token{
//		{token.TABLE, "table"},
//		{token.IDENT, "user"},
//		{token.LBRACE, "{"},
//		{token.IDENT, "user_id"},
//		{token.IDENT, "int"},
//		{token.IDENT, "notnull"},
//		{token.COMMA, ","},
//		{token.IDENT, "username"},
//		{token.IDENT, "string"},
//		{token.IDENT, "null"},
//		{token.COMMA, ","},
//		{token.IDENT, "age"},
//		{token.IDENT, "int"},
//		{token.IDENT, "min"},
//		{token.LPAREN, "("},
//		{token.INTEGER, "18"},
//		{token.RPAREN, ")"},
//		{token.COMMA, ","},
//		{token.IDENT, "gender"},
//		{token.IDENT, "bool"},
//		{token.RBRACE, "}"},
//	}
//
//	var s scanner.Scanner
//	s.Init([]byte(input))
//
//	for _, tt := range tests {
//		tok := s.Next()
//		if tok.Type != tt.Type {
//			t.Errorf("Next() type wrong, Tok = %q, want %q", tok.Type, tt.Type)
//		}
//		if tok.Literal != tt.Literal {
//			t.Errorf("Next() literal wrong, Lit = %v, want %v", tok, tt.Literal)
//		}
//	}
//
//	if s.ErrorCount != 0 {
//		t.Errorf("got %d errors", s.ErrorCount)
//	}
//
//}
