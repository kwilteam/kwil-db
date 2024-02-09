package sqlparser

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/kwilteam/sql-grammar-go/sqlgrammar"
)

// Parse parses a raw sql string and returns a tree.Ast
func Parse(sql string) (ast tree.Ast, err error) {
	currentLine := 1
	return ParseSql(sql, currentLine, nil, false)
}

// ParseSql parses a single raw sql statement and returns tree.Ast
func ParseSql(sql string, currentLine int, errorListener *ErrorListener, trace bool) (ast tree.Ast, err error) {
	var visitor *KFSqliteVisitor

	if errorListener == nil {
		errorListener = NewErrorListener()
	}

	stream := antlr.NewInputStream(sql)
	lexer := sqlgrammar.NewSQLiteLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := sqlgrammar.NewSQLiteParser(tokenStream)

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

	visitor = NewKFSqliteVisitor(KFVisitorWithTrace(trace))

	parseTree := p.Parse()
	result := visitor.Visit(parseTree)
	// since we only expect a single statement
	return result.([]tree.Ast)[0], err
}
