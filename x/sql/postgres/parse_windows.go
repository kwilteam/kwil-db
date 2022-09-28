//go:build windows
// +build windows

package postgres

import (
	"errors"
	"io"

	"kwil/x/sql/ast"
	"kwil/x/sql/core"
)

func NewParser() *Parser {
	return &Parser{}
}

type Parser struct {
}

func (p *Parser) Parse(r io.Reader) ([]ast.Statement, error) {
	return nil, core.ErrUnsupportedOS
}

// https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-SYNTAX-COMMENTS
func (p *Parser) CommentSyntax() core.CommentSyntax {
	return core.CommentSyntax{
		Dash:      true,
		SlashStar: true,
	}
}

func (p *Parser) IsReservedKeyword(string) bool {
	return false
}

func (p *Parser) Kind() core.EngineKind {
	return core.EnginePostgreSQL
}
