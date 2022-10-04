package engine

import (
	"fmt"

	"kwil/x/sql/catalog"
	"kwil/x/sql/sqlparse/ast"
)

type QueryCatalog struct {
	catalog *catalog.Catalog
	ctes    map[string]*Table
}

func buildQueryCatalog(c *catalog.Catalog, node ast.Node) (*QueryCatalog, error) {
	var with *ast.WithClause
	switch n := node.(type) {
	case *ast.DeleteStmt:
		with = n.WithClause
	case *ast.InsertStmt:
		with = n.WithClause
	case *ast.UpdateStmt:
		with = n.WithClause
	case *ast.SelectStmt:
		with = n.WithClause
	default:
		with = nil
	}
	qc := &QueryCatalog{catalog: c, ctes: map[string]*Table{}}
	if with != nil {
		for _, item := range with.Ctes.Items {
			if cte, ok := item.(*ast.CommonTableExpr); ok {
				cols, err := outputColumns(qc, cte.Ctequery)
				if err != nil {
					return nil, err
				}
				rel := &ast.TableName{Name: *cte.Ctename}
				for i := range cols {
					cols[i].Table = &catalog.QualName{Catalog: rel.Catalog, Schema: rel.Schema, Name: rel.Name}
				}
				qc.ctes[*cte.Ctename] = &Table{
					Rel:     &catalog.QualName{Catalog: rel.Catalog, Schema: rel.Schema, Name: rel.Name},
					Columns: cols,
				}
			}
		}
	}
	return qc, nil
}

func convertColumn(rel *ast.TableName, c *catalog.Column) *Column {
	return &Column{
		Table:    &catalog.QualName{Catalog: rel.Catalog, Schema: rel.Schema, Name: rel.Name},
		Name:     c.Name,
		DataType: dataType(c.Type),
		NotNull:  c.IsNotNull,
		IsArray:  c.IsArray,
		Type:     c.Type,
		Length:   c.Length,
	}
}

func (qc QueryCatalog) GetTable(rel *ast.TableName) (*Table, error) {
	cte, exists := qc.ctes[rel.Name]
	if exists {
		return cte, nil
	}
	src, ok := qc.catalog.Table(rel.Schema, rel.Name)
	if !ok {
		return nil, fmt.Errorf("table not found: %s", rel.Name)
	}
	var cols []*Column
	for _, c := range src.Columns {
		cols = append(cols, convertColumn(rel, c))
	}
	return &Table{Rel: &catalog.QualName{Catalog: rel.Catalog, Schema: rel.Schema, Name: rel.Name}, Columns: cols}, nil
}

func (qc QueryCatalog) GetFunc(rel *ast.FuncName) (*Function, error) {
	funcs, err := qc.catalog.Funcs(rel.Schema, rel.Name)
	if err != nil {
		return nil, err
	}
	if len(funcs) == 0 {
		return nil, fmt.Errorf("function not found: %s", rel.Name)
	}
	return &Function{
		Rel:        &catalog.QualName{Catalog: rel.Catalog, Schema: rel.Schema, Name: rel.Name},
		ReturnType: funcs[0].ReturnType,
	}, nil
}
