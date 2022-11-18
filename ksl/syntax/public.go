package syntax

import (
	"ksl"
	"ksl/syntax/ast"
	"ksl/syntax/lex"
	"os"
)

type ParseOption func(*parser)

func Parse(src []byte, filename string, start ksl.Pos, opts ...ParseOption) (*ast.File, ksl.Diagnostics) {
	tokens, diags := Lex(src, filename, start)

	peeker := newPeeker(tokens, false)
	parser := &parser{peeker: peeker}
	for _, opt := range opts {
		opt(parser)
	}

	file, parseDiags := parser.parseFile(src, filename)
	diags = append(diags, parseDiags...)

	parser.AssertEmptyIncludeNewlinesStack()

	return file, diags
}

func Lex(src []byte, filename string, start ksl.Pos) (lex.Tokens, ksl.Diagnostics) {
	tokens := lex.ScanTokens(src, filename, start)
	diags := lex.CheckInvalidTokens(tokens)
	return tokens, diags
}

func ParseFiles(filenames ...string) ([]*ast.File, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var files []*ast.File

	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		if err != nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Failed to read file",
				Detail:   err.Error(),
			})
		} else {
			file, fileDiags := Parse(data, filename, ksl.InitialPos)
			diags = append(diags, fileDiags...)
			files = append(files, file)
		}
	}
	return files, diags
}
