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

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/action-grammar-go/actgrammar"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
)

// Parse parses an action statement string and returns an ActionStmt.
// A new error listener will be created, and parsing trace is disabled.
func Parse(stmt string) (ast ActionStmt, err error) {
	return ParseActionStmt(stmt, nil, false, false)
}

// ParseActionStmt parses a single action statement and returns an ActionStmt.
// errorListener is optional, if nil, a new error listener is created, it's
// mostly used for testing.
// trace is optional, if true, parsing trace will be enabled.
func ParseActionStmt(stmt string, errorListener *sqlparser.ErrorListener,
	trace bool, trackPos bool) (ast ActionStmt, err error) {
	var visitor *astBuilder

	if errorListener == nil {
		errorListener = sqlparser.NewErrorListener()
	}

	stream := antlr.NewInputStream(stmt)
	lexer := actgrammar.NewActionLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := actgrammar.NewActionParser(tokenStream)

	// remove default error visitor
	p.RemoveErrorListeners()
	p.AddErrorListener(errorListener)

	p.BuildParseTrees = true

	defer func() {
		if e := recover(); e != nil {
			errorListener.Add(fmt.Sprintf("panic: %v", e))
		}

		if err != nil {
			errorListener.AddError(err)
		}

		err = errorListener.Err()
	}()

	visitor = newAstBuilder(trace, trackPos)

	parseTree := p.Statement()
	result := visitor.Visit(parseTree)
	return result.([]ActionStmt)[0], err
}
