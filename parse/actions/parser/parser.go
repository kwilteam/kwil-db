// Package actparser contains the parser for the statements inside the kuneiform
// action block.
// This package is only temporary, to reduce the need to change our public
// kuneiform schema. Once our schema is stable, we will remove this package, and
// put actual Stmt types to kuneiform schema (so engine doesn't need to parse).
//
// By having this package, we can just check the syntax of the action block
// without parsing, then pass the whole statement to the engine. The engine will
// parse the statements to its needs.
package actparser

import (
	"fmt"
	"runtime"

	"github.com/antlr4-go/antlr/v4"

	actgrammar "github.com/kwilteam/kwil-db/parse/actions/gen"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

// Parse parses multiple action statements and returns a slice of ActionStmt.
// This is to maintain compatibility with the old function signature.
func Parse(stmt string, errLis parseTypes.AntlrErrorListener) (asts []ActionStmt, err error) {
	var visitor *astBuilder

	stream := antlr.NewInputStream(stmt)
	lexer := actgrammar.NewActionLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := actgrammar.NewActionParser(tokenStream)

	// remove default error visitor
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errLis)
	p.RemoveErrorListeners()
	p.AddErrorListener(errLis)

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
			if errLis.Err() != nil {
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

	visitor = &astBuilder{
		errs: errLis,
	}

	parseTree := p.Statement()
	result := visitor.Visit(parseTree)
	return result.([]ActionStmt), err
}
