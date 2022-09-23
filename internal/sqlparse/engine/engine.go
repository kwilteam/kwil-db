package engine

import (
	"io"

	"github.com/kwilteam/kwil-db/internal/sqlparse/ast"
)

type Parser interface {
	Parse(io.Reader) ([]ast.Statement, error)
	CommentSyntax() CommentSyntax
	IsReservedKeyword(string) bool
}

type CommentSyntax struct {
	Dash      bool
	Hash      bool
	SlashStar bool
}
