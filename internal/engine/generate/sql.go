package generate

import (
	"fmt"

	"github.com/kwilteam/kwil-db/parse"
)

// WriteSQL converts a SQL node to a string.
// It can optionally rewrite named parameters to numbered parameters.
// If so, it returns the order of the parameters in the order they appear in the statement.
func WriteSQL(node *parse.SQLStatement, orderParams bool, pgSchema string) (stmt string, params []string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
		}
	}()
	sqlGen := &sqlGenerator{
		pgSchema: pgSchema,
	}
	if orderParams {
		rewriter := &sqlParamRewriter{
			sqlGenerator: *sqlGen,
		}
		stmt = node.Accept(rewriter).(string)
		params = rewriter.paramOrder
	} else {
		stmt = node.Accept(sqlGen).(string)
	}

	return stmt + ";", params, nil
}

// sqlParamRewriter is a rewriter that rewrites SQL parameters to placeholders.
type sqlParamRewriter struct {
	sqlGenerator
	// paramOrder is the name of the parameters in the order they appear in the action.
	// Since the actionGenerator rewrites actions from named parameters to numbered parameters,
	// the order of the named parameters is stored here.
	paramOrder []string
}

func (a *sqlParamRewriter) VisitExpressionVariable(p0 *parse.ExpressionVariable) any {
	str := p0.String()

	// if it already exists, we write it as that index. otherwise, we add it to the list.
	// Postgres uses $1, $2, etc. for numbered parameters.
	for i, v := range a.paramOrder {
		if v == str {
			return "$" + fmt.Sprint(i+1)
		}
	}

	a.paramOrder = append(a.paramOrder, str)
	return "$" + fmt.Sprint(len(a.paramOrder))
}
