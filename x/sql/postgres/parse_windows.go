//go:build windows
// +build windows

package postgres

import (
	"io"

	"kwil/x/sql/ast"
	"kwil/x/sql/core"
)

func (p *Parser) Parse(r io.Reader) ([]ast.Statement, error) {
	return nil, core.ErrUnsupportedOS
}
