package engine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kwilteam/kwil-db/internal/sql/ast"
	"github.com/kwilteam/kwil-db/internal/sql/catalog"
	"github.com/kwilteam/kwil-db/internal/sql/core"
	"github.com/kwilteam/kwil-db/internal/sql/source"
	"github.com/kwilteam/kwil-db/internal/sql/sqlerr"
	sqlx "github.com/kwilteam/kwil-db/internal/sql/x"
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
	catalog.Updater
	*catalog.Catalog
}

func NewEngine(parser SQLParser, catalog *catalog.Catalog, updater catalog.Updater) *Engine {
	return &Engine{
		SQLParser: parser,
		Updater:   updater,
		Catalog:   catalog,
	}
}

func (c *Engine) ParseDDLFiles(schemas []string) error {
	files, err := sqlx.Glob(schemas)
	if err != nil {
		return err
	}
	var e error
	for _, filename := range files {
		blob, err := os.ReadFile(filename)
		if err != nil {
			e = multierr.Append(e, err)
			continue
		}
		contents := string(blob)
		if err := c.ParseDDL(contents); err != nil {
			e = multierr.Append(e, err)
			continue
		}
	}

	return e
}

func (c *Engine) ParseDDL(src string) error {
	var e error
	stmts, err := c.Parse(strings.NewReader(src))
	if err != nil {
		return err
	}
	for i := range stmts {
		if err := c.UpdateDDL(stmts[i], &colOutputter{c.Catalog}); err != nil {
			e = multierr.Append(e, source.NewError("", src, stmts[i].Raw.Pos(), err))
			continue
		}
	}
	return e
}

func (c *Engine) ParseQueryFiles(paths ...string) ([]*Query, error) {
	var q []*Query
	var errs []error

	files, err := sqlx.Glob(paths)
	if err != nil {
		return nil, err
	}

	for _, filename := range files {
		blob, err := os.ReadFile(filename)
		if err != nil {
			errs = append(errs, source.NewError(filename, "", 0, err))
			continue
		}

		src := string(blob)
		queries, err := c.ParseStatement(src)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		q = append(q, queries...)
	}

	if len(q) == 0 {
		return nil, fmt.Errorf("no queries contained in paths %s", strings.Join(paths, ","))
	}

	return q, multierr.Combine(errs...)
}

func (c *Engine) ParseStatement(src string) ([]*Query, error) {
	var q []*Query
	var errs []error
	if !strings.HasSuffix(src, ";") {
		src += ";"
	}

	stmts, err := c.Parse(strings.NewReader(src))
	if err != nil {
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
