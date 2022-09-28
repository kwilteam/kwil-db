package engine

import (
	"errors"
	"io"
	"strings"

	"kwil/x/sql/ast"
	"kwil/x/sql/catalog"
	"kwil/x/sql/core"
	"kwil/x/sql/source"
	"kwil/x/sql/sqlerr"

	"go.uber.org/multierr"
)

type SQLParser interface {
	Parse(io.Reader) ([]ast.Statement, error)
	CommentSyntax() core.CommentSyntax
	IsReservedKeyword(string) bool
	Kind() core.EngineKind
}

type Engine struct {
	SQLParser
	*catalog.Catalog
}

func NewEngine(parser SQLParser, catalog *catalog.Catalog) *Engine {
	return &Engine{
		SQLParser: parser,
		Catalog:   catalog,
	}
}

func (c *Engine) ParseStatement(src string) ([]*Query, error) {
	var q []*Query
	var errs []error
	if !strings.HasSuffix(src, ";") {
		src += ";"
	}

	stmts, err := c.Parse(strings.NewReader(src))
	if err != nil {
		if errors.Is(err, core.ErrUnsupportedOS) {
			return nil, err
		}
		return nil, source.NewError("", src, 0, err)
	}

	for _, stmt := range stmts {
		query, err := c.parseQuery(stmt.Raw, src)
		if err == ErrUnsupportedStatementType {
			continue
		}
		if err != nil {
			var e *sqlerr.Error
			loc := stmt.Raw.Pos()
			if errors.As(err, &e) && e.Location != 0 {
				loc = e.Location
			}
			errs = append(errs, source.NewError("", src, loc, err))
			continue
		}
		if query != nil {
			q = append(q, query)
		}
	}

	return q, multierr.Combine(errs...)
}
