package parser

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/parse/procedures/gen"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
)

func Parse(stmt string) ([]Statement, error) {
	return ParseWithErrorListener(stmt, nil)
}

func ParseWithErrorListener(stmt string, errorListener *sqlparser.ErrorListener) (clauses []Statement, err error) {
	visitor := &proceduralLangVisitor{}

	if errorListener == nil {
		errorListener = sqlparser.NewErrorListener()
	}

	stream := antlr.NewInputStream(stmt)
	lexer := gen.NewProcedureLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := gen.NewProcedureParser(tokenStream)

	if errorListener != nil {
		// remove default error visitor
		p.RemoveErrorListeners()
		p.AddErrorListener(errorListener)
	}

	p.BuildParseTrees = true

	defer func() {
		if e := recover(); e != nil {
			errorListener.Add(fmt.Sprintf("%v", e))
		}

		if err != nil {
			errorListener.AddError(err)
		}

		err = errorListener.Err()
	}()

	result := visitor.Visit(p.Program())

	if errorListener.Err() != nil {
		return nil, errorListener.Err()
	}

	res, ok := result.([]Statement)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return res, nil
}
