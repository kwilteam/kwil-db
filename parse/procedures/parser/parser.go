package parser

import (
	"fmt"
	"runtime"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/parse/procedures/gen"
	"github.com/kwilteam/kwil-db/parse/types"
)

// Parse should only be used when the caller does not care about getting
// in-depth error information.
func Parse(stmt string) ([]Statement, error) {
	errLis := types.NewErrorListener()
	res, err := ParseWithErrorListener(stmt, errLis)
	if err != nil {
		return nil, err
	}
	if errLis.Err() != nil {
		return nil, errLis.Err()
	}

	return res, nil
}

// ParseOpts are options for parsing a procedural language statement.
type ParseOpts struct {
	// ErrorListener is the error listener to use when parsing the statement.
	// If not provided, it will default to a new error listener, and all
	// parsing errors will be returned as an error.
	ErrorListener types.AntlrErrorListener
}

// ParseWithErrorListener parses a procedural language statement and returns the AST.
func ParseWithErrorListener(stmt string, errorListener types.AntlrErrorListener) (clauses []Statement, err error) {
	visitor := &proceduralLangVisitor{
		errs: errorListener,
	}

	stream := antlr.NewInputStream(stmt)
	lexer := gen.NewProcedureLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := gen.NewProcedureParser(tokenStream)

	// remove default error listeners
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

	result := visitor.Visit(p.Program())

	res, ok := result.([]Statement)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return res, nil
}
