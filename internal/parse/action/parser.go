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

	"github.com/antlr4-go/antlr/v4"

	"github.com/kwilteam/action-grammar-go/actgrammar"
	sqlparser "github.com/kwilteam/kwil-db/internal/parse/sql"
)

// Parse parses multiple action statements and returns a slice of ActionStmt.
// This is to maintain compatibility with the old function signature.
// TODO: Remove this function and use Parse instead.
func Parse(stmt string) (asts []ActionStmt, err error) {
	var visitor *astBuilder

	errorListener := sqlparser.NewErrorListener()

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

	visitor = newAstBuilder(false, false)

	parseTree := p.Statement()
	result := visitor.Visit(parseTree)
	return result.([]ActionStmt), err
}
