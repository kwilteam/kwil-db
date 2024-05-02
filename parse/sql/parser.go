package sqlparser

import (
	"fmt"
	"runtime"

	antlr "github.com/antlr4-go/antlr/v4"

	grammar "github.com/kwilteam/kwil-db/parse/sql/gen"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

// Parse parses a raw sql string and returns a tree.Statement.
// It should only be used if the user is trying to parse a single
// sql statement, and doesn't care about error handling.
func Parse(sql string) (tree.Statement, error) {
	stmts, err := ParseMany(sql)
	if err != nil {
		return nil, err
	}

	if len(stmts) != 1 {
		return nil, fmt.Errorf("expected 1 statement, but found %d", len(stmts))
	}

	return stmts[0], nil
}

// ParseMany parses a raw sql string and returns tree.Statements.
// It should only be used if the user is trying to parse multiple
// sql statements, and doesn't care about error handling.
func ParseMany(sql string) ([]tree.Statement, error) {
	errorListener := parseTypes.NewErrorListener()
	ast, err := ParseWithErrorListener(sql, errorListener)
	if err != nil {
		return nil, err
	}

	if errorListener.Err() != nil {
		return nil, errorListener.Err()
	}

	return ast, nil
}

// ParseWithErrorListener parses a raw sql string and returns tree.Statements.
// Syntax errors are returned in the error listener.
func ParseWithErrorListener(sql string, errorListener parseTypes.AntlrErrorListener) (stmts []tree.Statement, err error) {
	stream := antlr.NewInputStream(sql)
	lexer := grammar.NewSQLLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := grammar.NewSQLParser(tokenStream)

	// remove default error visitor
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errorListener)

	p.RemoveErrorListeners()
	p.AddErrorListener(errorListener)

	p.BuildParseTrees = true

	defer func() {
		if e := recover(); e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", e)
			}

			// if there is a panic, it is likely due to a syntax error
			// check for parse errors and return them first
			if errorListener.Err() != nil {
				// if there is an error listener error, we should swallow the panic
				// If the issue persists until after the user has fixed the parse errors,
				// the panic will be returned in the else block.
				err = nil
			} else {
				// if there are no parse errors, then there is a bug.
				// we should return the panic with a stack trace.
				buf := make([]byte, 1<<16)
				stackSize := runtime.Stack(buf, false)

				err = fmt.Errorf("%w\n\n%s", err, buf[:stackSize])
			}
		}
	}()

	visitor := &astBuilder{
		errs: errorListener,
	}
	parsed := p.Statements().Accept(visitor).([]tree.AstNode)

	s := make([]tree.Statement, len(parsed))
	for i, stmt := range parsed {
		s[i] = stmt.(tree.Statement)
	}

	return s, nil
}
