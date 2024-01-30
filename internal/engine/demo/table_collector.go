package demo

import "github.com/kwilteam/kwil-db/parse/sql/tree"

// tableCollector collects all the table names in a statement
type tableCollector struct {
	*tree.BaseAstVisitor

	tables map[string]*tree.TableOrSubqueryTable
	//cteSchemas []*schema
}

func newTableCollector() *tableCollector {
	return &tableCollector{
		tables: make(map[string]*tree.TableOrSubqueryTable),
	}
}

func (d *tableCollector) collect(node tree.AstNode) map[string]*tree.TableOrSubqueryTable {
	d.Visit(node)
	return d.tables
}

func (d *tableCollector) VisitSelect(node *tree.Select) any {
	// first resolve the CTEs, generate temp types.Table
	// then, collect all refered schemas
	//if node.CTE != nil {
	//	for _, cte := range node.CTE {
	//		cte.Select.Accept(d)
	//		d.VisitSelectStmt(cte.Select)
	//	}
	//}

	d.VisitSelectStmt(node.SelectStmt)
	return nil
}

func (d *tableCollector) VisitSelectStmt(node *tree.SelectStmt) any {
	for _, core := range node.SelectCores {
		core.Accept(d)
		d.VisitSelectCore(core)
	}

	return nil
}

func (d *tableCollector) VisitSelectCore(node *tree.SelectCore) any {
	if node.From != nil {
		d.VisitTableOrSubquery(node.From.JoinClause.TableOrSubquery)
		for _, join := range node.From.JoinClause.Joins {
			d.VisitTableOrSubquery(join.Table)
		}
	}
	return nil
}

func (d *tableCollector) VisitTableOrSubqueryTable(node *tree.TableOrSubqueryTable) any {
	if node.Alias != "" {
		d.tables[node.Alias] = node
	} else {
		d.tables[node.Name] = node
	}
	return nil
}
