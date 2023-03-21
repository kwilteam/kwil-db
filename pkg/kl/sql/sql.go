package sql

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"kwil/internal/pkg/sqlite"
)

func ParseRawSQL(sql string, currentLine int, queryOnly bool) (err error) {
	KlSQLInit()

	eh := &errorHandler{CurLine: currentLine}
	listener := NewSqliteListener(eh)
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
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
