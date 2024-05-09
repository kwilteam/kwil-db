package tree

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/parse/types"
)

// AstNode represents an AST node.
type AstNode interface {
	AstWalker
	GetNode() *types.Node
	Set(rule antlr.ParserRuleContext)
	SetToken(tok antlr.Token)

	// ToSQL converts the node to a SQL string.
	ToSQL() string
	// Accept accepts an AstVisitor to visit itself.
	Accept(AstVisitor) any
}

type Statement interface {
	AstNode

	statement()
}

type Expression interface {
	AstNode
	joinable
}

type ResultColumn interface {
	AstNode

	resultColumn()
}
