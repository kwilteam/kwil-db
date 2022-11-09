package kslsyntax

import (
	"ksl"
	"ksl/kslsyntax/ast"
	"ksl/kslsyntax/lex"
)

type ParseOption func(*parser)

func Parse(src []byte, filename string, start ksl.Pos, opts ...ParseOption) (*ast.Document, ksl.Diagnostics) {
	tokens, diags := Lex(src, filename, start)

	peeker := newPeeker(tokens, false)
	parser := &parser{peeker: peeker}
	for _, opt := range opts {
		opt(parser)
	}

	doc, parseDiags := parser.parseDocument()
	diags = append(diags, parseDiags...)

	parser.AssertEmptyIncludeNewlinesStack()

	return doc, diags
}

func Lex(src []byte, filename string, start ksl.Pos) (lex.Tokens, ksl.Diagnostics) {
	tokens := lex.ScanTokens(src, filename, start)
	diags := lex.CheckInvalidTokens(tokens)
	return tokens, diags
}
