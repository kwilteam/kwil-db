package postgres

import (
	"kwil/x/sql/core"
	"strings"
)

func NewParser() *Parser {
	return &Parser{}
}

type Parser struct {
}

// https://www.postgresql.org/docs/current/sql-keywords-appendix.html
func (p *Parser) IsReservedKeyword(s string) bool {
	switch strings.ToLower(s) {
	case "all":
	case "analyse":
	case "analyze":
	case "and":
	case "any":
	case "array":
	case "as":
	case "asc":
	case "asymmetric":
	case "authorization":
	case "binary":
	case "both":
	case "case":
	case "cast":
	case "check":
	case "collate":
	case "collation":
	case "column":
	case "concurrently":
	case "constraint":
	case "create":
	case "cross":
	case "current_catalog":
	case "current_date":
	case "current_role":
	case "current_schema":
	case "current_time":
	case "current_timestamp":
	case "current_user":
	case "default":
	case "deferrable":
	case "desc":
	case "distinct":
	case "do":
	case "else":
	case "end":
	case "except":
	case "false":
	case "fetch":
	case "for":
	case "foreign":
	case "freeze":
	case "from":
	case "full":
	case "grant":
	case "group":
	case "having":
	case "ilike":
	case "in":
	case "initially":
	case "inner":
	case "intersect":
	case "into":
	case "is":
	case "isnull":
	case "join":
	case "lateral":
	case "leading":
	case "left":
	case "like":
	case "limit":
	case "localtime":
	case "localtimestamp":
	case "natural":
	case "not":
	case "notnull":
	case "null":
	case "offset":
	case "on":
	case "only":
	case "or":
	case "order":
	case "outer":
	case "overlaps":
	case "placing":
	case "primary":
	case "references":
	case "returning":
	case "right":
	case "select":
	case "session_user":
	case "similar":
	case "some":
	case "symmetric":
	case "table":
	case "tablesample":
	case "then":
	case "to":
	case "trailing":
	case "true":
	case "union":
	case "unique":
	case "user":
	case "using":
	case "variadic":
	case "verbose":
	case "when":
	case "where":
	case "window":
	case "with":
	default:
		return false
	}
	return true
}

// https://www.postgresql.org/docs/current/sql-syntax-lexical.html#SQL-SYNTAX-COMMENTS
func (p *Parser) CommentSyntax() core.CommentSyntax {
	return core.CommentSyntax{
		Dash:      true,
		SlashStar: true,
	}
}

func (p *Parser) Kind() core.EngineKind {
	return core.EnginePostgreSQL
}
