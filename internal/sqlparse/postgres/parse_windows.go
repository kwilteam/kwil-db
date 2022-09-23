//go:build windows
// +build windows

package postgres

import (
	"errors"
	"io"

	"github.com/kwilteam/kwil-db/internal/sqlparse/ast"
	"github.com/kwilteam/kwil-db/internal/sqlparse/metadata"
)

func NewParser() *Parser {
	return &Parser{}
}

type Parser struct {
}

func (p *Parser) Parse(r io.Reader) ([]ast.Statement, error) {
	return nil, errors.New("the PostgreSQL engine does not support Windows")
}

// https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-SYNTAX-COMMENTS
func (p *Parser) CommentSyntax() metadata.CommentSyntax {
	return metadata.CommentSyntax{
		Dash:      true,
		SlashStar: true,
	}
}
