package sql_parser

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
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
