package datasets

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/sql_parser/sqlite"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

func parseAliases(sql string) (map[string]string, error) {
	stream := antlr.NewInputStream(sql)
	lexer := sqlite.NewSQLiteLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := sqlite.NewSQLiteParser(tokenStream)

	el := new(antlr.DefaultErrorListener)

	p.AddErrorListener(el)

	p.BuildParseTrees = true

	tree := p.Parse()

	ap := newAliasParser()

	antlr.ParseTreeWalkerDefault.Walk(ap, tree)

	if len(ap.errs) > 0 {
		return nil, ap.errs[0]
	}

	return ap.aliases, nil
}

func newAliasParser() *AliasParser {
	return &AliasParser{
		aliases: make(map[string]string),
	}
}

type AliasParser struct {
	*sqlite.BaseSQLiteParserListener

	aliases map[string]string
	errs    []error
}

func (a *AliasParser) EnterTable_name(ctx *sqlite.Table_nameContext) {
	tableName := ctx.GetText()

	parentCtx, ok := ctx.GetParent().(antlr.RuleContext)
	if !ok {
		a.errs = append(a.errs, fmt.Errorf("error parsing table names and aliases"))
		return
	}

	// Check if the parent context is Table_or_subqueryContext
	if tosCtx, ok := parentCtx.(*sqlite.Table_or_subqueryContext); ok {
		// Look for an alias in the Table_or_subqueryContext
		if aliasCtx := tosCtx.Table_alias(); aliasCtx != nil {
			alias := aliasCtx.GetText()
			a.aliases[alias] = tableName
		}
	}
}
