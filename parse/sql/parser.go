package sqlparser

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"

	"github.com/kwilteam/kwil-db/parse/sql/grammar"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// Parse parses a raw sql string and returns a tree.Statement
func Parse(sql string) (ast tree.Statement, err error) {
	currentLine := 1
	return ParseSql(sql, currentLine, nil, false, false)
}

// ParseSql parses a single raw sql statement and returns tree.Statement
func ParseSql(sql string, currentLine int, errorListener *ErrorListener,
	trace bool, withPos bool) (ast tree.Statement, err error) {
	var visitor *astBuilder

	if errorListener == nil {
		errorListener = NewErrorListener()
	}

	stream := antlr.NewInputStream(sql)
	lexer := grammar.NewSQLLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := grammar.NewSQLParser(tokenStream)

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

	visitor = newAstBuilder(astBuilderWithTrace(trace), astBuilderWithPos(withPos))
	stmts := p.Statements().Accept(visitor).([]tree.AstNode)
	// since we only expect a single statement
	return stmts[0].(tree.Statement), err
}
