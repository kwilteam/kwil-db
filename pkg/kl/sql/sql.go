package sql

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"kwil/internal/pkg/kl/types"
	"kwil/internal/pkg/sqlite"
	"kwil/pkg/kl/token"
)

func ParseRawSQL(sql string, currentLine int, actionName string, dbCtx types.DatabaseContext, trace bool) (err error) {
	KlSQLInit()

	var listener *KlSqliteListener
	eh := &errorHandler{CurLine: currentLine}
	if trace {
		listener = NewKlSqliteListener(eh, actionName, dbCtx, WithTrace())
	} else {
		listener = NewKlSqliteListener(eh, actionName, dbCtx)
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
			listener.Errors.Add(token.Position{0, 0}, err.Error())
		}

		err = listener.Errors.Err()
	}()

	stream := antlr.NewInputStream(sql)
	lexer := sqlite.NewSQLiteLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := sqlite.NewSQLiteParser(tokenStream)

	el := new(sqliteErrorListener)
	p.AddErrorListener(el)

	p.BuildParseTrees = true

	//// execute during parsing(careful don't mess up parser throwing error)
	//p.AddParseListener(listener)
	//p.Parse()
	// or after parsing, execute while walking the tree
	tree := p.Parse()
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	return err
}
