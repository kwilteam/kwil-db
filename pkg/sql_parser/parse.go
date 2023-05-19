package sql_parser

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/kwil-db/pkg/engine/tree"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/token"
	"github.com/kwilteam/kwil-db/pkg/sql_parser/sqlite"
)

func ParseRawSQL(sql string, currentLine int, actionName string, dbCtx DatabaseContext,
	errorListener *sqliteErrorListener, trace bool, walkTree bool) (err error) {
	KlSQLInit()

	var listener *KlSqliteListener

	if errorListener == nil {
		errorHandler := NewErrorHandler(currentLine)
		errorListener = newSqliteErrorListener(errorHandler)
	}

	if trace {
		listener = NewKlSqliteListener(errorListener.ErrorHandler, actionName, dbCtx, WithTrace())
	} else {
		listener = NewKlSqliteListener(errorListener.ErrorHandler, actionName, dbCtx)
	}

	stream := antlr.NewInputStream(sql)
	lexer := sqlite.NewSQLiteLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := sqlite.NewSQLiteParser(tokenStream)

	// remove default error listener
	p.RemoveErrorListeners()
	p.AddErrorListener(errorListener)

	p.BuildParseTrees = true

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
			listener.Errors.Add(token.Position{0, 0}, err.Error())
		}

		err = listener.Errors.Err()
	}()

	//// execute during parsing(careful don't mess up parser_inner throwing error)
	//p.AddParseListener(listener)
	//p.Parse()
	// or after parsing, execute while walking the tree
	tree := p.Parse()

	if walkTree {
		antlr.ParseTreeWalkerDefault.Walk(listener, tree)
	}

	return err
}

func ParseRawSQLVisitor(sql string, currentLine int, actionName string, dbCtx DatabaseContext,
	errorListener *sqliteErrorListener, trace bool, walkTree bool) (ast tree.Ast, err error) {
	KlSQLInit()
	var visitor *KFSqliteVisitor

	if errorListener == nil {
		errorHandler := NewErrorHandler(currentLine)
		errorListener = newSqliteErrorListener(errorHandler)
	}

	stream := antlr.NewInputStream(sql)
	lexer := sqlite.NewSQLiteLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := sqlite.NewSQLiteParser(tokenStream)

	// remove default error visitor
	p.RemoveErrorListeners()
	p.AddErrorListener(errorListener)

	p.BuildParseTrees = true

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
			visitor.Errors.Add(token.Position{}, err.Error())
		}

		err = visitor.Errors.Err()
	}()

	//// execute during parsing(careful don't mess up parser_inner throwing error)
	//p.AddParseListener(visitor)
	//p.Parse()
	// or after parsing, execute while walking the tree
	parseCtx := p.Parse()

	if trace {
		visitor = NewKFSqliteVisitor(errorListener.ErrorHandler, actionName, dbCtx)
	} else {
		visitor = NewKFSqliteVisitor(errorListener.ErrorHandler, actionName, dbCtx)
	}

	result := visitor.Visit(parseCtx)
	return result.(tree.Ast), err
}
