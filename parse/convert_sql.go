package parse

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/parse/gen"
)

// this file converts the AST to our old SQL AST

type sqlConverter struct {
	antlr.BaseParseTreeVisitor
}

var _ gen.KuneiformParserVisitor = (*sqlConverter)(nil)

func (s *sqlConverter) VisitSql(ctx *gen.SqlContext) any {
	return s.VisitChildren(ctx)
}
