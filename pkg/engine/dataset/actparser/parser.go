package actparser

import (
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/action-grammar-go/actgrammar"
)

// Parse parses a action statement string and returns a ActionStmt
func Parse(stmt string) (ast ActionStmt, err error) {
	return ParseActionStmt(stmt, nil, false)
}

// ParseActionStmt parses a single action statement and returns
func ParseActionStmt(stmt string, errorListener *sqlparser.ErrorListener, trace bool) (ast ActionStmt, err error) {
	var visitor *KFActionVisitor

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
			errorListener.Add(err.Error())
		}

		err = errorListener.Err()
	}()

	visitor = NewKFActionVisitor(KFActionVisitorWithTrace(trace))

	parseTree := p.Statement()
	result := visitor.Visit(parseTree)
	return result.([]ActionStmt)[0], err
}
