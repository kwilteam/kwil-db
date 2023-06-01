package sql_parser

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/kwil-db/pkg/engine/tree"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/token"
	grammar "github.com/kwilteam/kwil-db/pkg/sql_parser/sql_grammar"
)

// Parse parses a raw sql string and returns a tree.Ast
func Parse(sql string) (ast tree.Ast, err error) {
	currentLine := 1
	errorHandler := NewErrorHandler(currentLine)
	errorListener := newSqliteErrorListener(errorHandler)

	return ParseSqlStatement(sql, currentLine, errorListener, false)
}

// ParseSqlStatement parses a single raw sql statement and returns tree.Ast
func ParseSqlStatement(sql string, currentLine int, errorListener *sqliteErrorListener, trace bool) (ast tree.Ast, err error) {
	var visitor *KFSqliteVisitor

	if errorListener == nil {
		errorHandler := NewErrorHandler(currentLine)
		errorListener = newSqliteErrorListener(errorHandler)
	}

	stream := antlr.NewInputStream(sql)
	lexer := grammar.NewSQLiteLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := grammar.NewSQLiteParser(tokenStream)

	// remove default error visitor
	p.RemoveErrorListeners()
	p.AddErrorListener(errorListener)

	p.BuildParseTrees = true

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
			errorListener.Errors.Add(token.Position{}, err.Error())
		}

		err = errorListener.Errors.Err()
	}()

	if trace {
		visitor = NewKFSqliteVisitor(errorListener.ErrorHandler, KFVisitorWithTrace())
	} else {
		visitor = NewKFSqliteVisitor(errorListener.ErrorHandler)
	}

	parseCtx := p.Parse()
	if errorListener.Errors.Err() != nil {
		return nil, errorListener.Errors.Err()
	}

	result := visitor.Visit(parseCtx)
	// since we only expect a single statement
	return result.([]tree.Ast)[0], err
}
