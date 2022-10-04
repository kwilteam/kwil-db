//go:build windows
// +build windows

package postgres

import (
	"errors"
	"io"

	"kwil/x/sql/sqlparse/ast"
	"kwil/x/sql/sqlparse/core"
)

func (p *Parser) Parse(r io.Reader) ([]ast.Statement, error) {
	return nil, core.ErrUnsupportedOS
}
