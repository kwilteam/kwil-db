/*
Package attributes analyzes a returned relations attributes, maintaining order.
This is useful for determining the relation schema that a query / CTE returns.

For example, given the following query:

	WITH satoshi_posts AS (
		SELECT id, title, content FROM posts
		WHERE user_id = (
			SELECT id FROM users WHERE username = 'satoshi' LIMIT 1
		)
	)
	SELECT id, title FROM satoshi_posts;

The attributes package will be able to determine that:
 1. The result of this query is a relation with two attributes: id and title
 2. The result of the common table expression satoshi_posts is a relation with three attributes: id, title, and content
*/
package attributes

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// A RelationAttribute is a column or expression that is part of a relation.
// It contains the logical representation of the attribute, as well as the
// data type
type RelationAttribute struct {
	// ResultExpression is the expression that represents the attribute
	// This can be things like "column_name", "table"."column_name", "sum(column_name)"", "5", etc.
	ResultExpression *tree.ResultColumnExpression

	// Type is the data type of the attribute
	Type types.DataType
}

// GetSelectCoreRelationAttributes will analyze the select core and return the
// identified relation.
// It returns a list of result column expressions.  These can be things like:
// tbl1.col, col, col AS alias, col*5 AS alias, etc.
// If a statement has "SELECT * FROM tbl",
// then the result column expressions will be tbl.col_1, tbl.col_2, etc.
func GetSelectCoreRelationAttributes(selectCore *tree.SelectCore, tables []*types.Table) ([]*RelationAttribute, error) {
	walker := newSelectCoreWalker(tables)
	err := selectCore.Walk(walker)
	if err != nil {
		return nil, fmt.Errorf("error analyzing select core: %w", err)
	}

	return walker.detectedAttributes, nil
}

func newSelectCoreWalker(tables []*types.Table) *selectCoreAnalyzer {
	return &selectCoreAnalyzer{
		AstWalker:          tree.NewBaseWalker(),
		context:            newSelectCoreContext(nil),
		schemaTables:       tables,
		detectedAttributes: []*RelationAttribute{},
	}
}

// selectCoreAnalyzer will walk the tree and identify the returned attributes for the select core
type selectCoreAnalyzer struct {
	tree.AstWalker
	context      *selectCoreContext
	schemaTables []*types.Table

	// detectedAttributes is a list of the detected attributes
	// from the scope
	detectedAttributes []*RelationAttribute
}

// newScope creates a new scope for the select core
// it sets the parent scope to the current scope
func (s *selectCoreAnalyzer) newScope() {
	oldCtx := s.context
	s.context = newSelectCoreContext(oldCtx)
}

// oldScope pops the current scope and returns to the parent scope
// if there is no parent scope, it simply sets the current scope to nil
func (s *selectCoreAnalyzer) oldScope() {
	if s.context == nil {
		panic("oldScope called with no current scope")
	}
	if s.context.parent == nil {
		s.context = nil
		return
	}

	s.context = s.context.parent
}

type selectCoreContext struct {
	// Parent is the parent context
	parent *selectCoreContext

	// results is the ordered list of query results
	results []tree.ResultColumn

	// usedTables is a list of tables used in the select core
	usedTables []*types.Table
}

// relations returns the identified relations
// it will expand the stars and table stars to the list of columns
func (s *selectCoreContext) relations() ([]*RelationAttribute, error) {
	results := make([]*RelationAttribute, 0)

	for _, res := range s.results {
		exprs, err := s.evaluateResult(res)
		if err != nil {
			return nil, err
		}

		results = append(results, exprs...)
	}

	return results, nil
}

// addResult adds a result to the list of results
func (s *selectCoreContext) addResult(result tree.ResultColumn) {
	s.results = append(s.results, result)
}

// evaluateResult evaluates a result, returning it as a column expression
func (s *selectCoreContext) evaluateResult(result tree.ResultColumn) ([]*RelationAttribute, error) {
	results := []*RelationAttribute{}

	switch r := result.(type) {
	default:
		panic(fmt.Sprintf("unknown result type: %T", r))
	case *tree.ResultColumnExpression:
		copied := *r

		dataType, err := predictReturnType(r.Expression, s.usedTables)
		if err != nil {
			return nil, err
		}

		if len(s.usedTables) > 0 {
			err := addTableIfNotPresent(s.usedTables[0].Name, &copied)
			if err != nil {
				return nil, err
			}
		}

		results = append(results, &RelationAttribute{
			ResultExpression: &copied,
			Type:             dataType,
		})
	case *tree.ResultColumnStar:
		for _, tbl := range s.usedTables {
			for _, col := range tbl.Columns {
				results = append(results, &RelationAttribute{
					ResultExpression: &tree.ResultColumnExpression{
						Expression: &tree.ExpressionColumn{
							Table:  tbl.Name,
							Column: col.Name,
						},
					},
					Type: col.Type,
				})
			}
		}
	case *tree.ResultColumnTable:
		tbl, err := findTable(s.usedTables, r.TableName)
		if err != nil {
			return nil, err
		}

		for _, col := range tbl.Columns {
			results = append(results, &RelationAttribute{
				ResultExpression: &tree.ResultColumnExpression{
					Expression: &tree.ExpressionColumn{
						Table:  tbl.Name,
						Column: col.Name,
					},
				},
				Type: col.Type,
			})
		}
	}

	return results, nil
}

// newSelectCoreContext creates a new select core context
func newSelectCoreContext(parent *selectCoreContext) *selectCoreContext {
	return &selectCoreContext{
		parent: parent,
	}
}

// EnterSelectCore creates a new scope.
func (s *selectCoreAnalyzer) EnterSelectCore(node *tree.SelectCore) error {
	s.newScope()

	return nil
}

// ExitSelectCore pops the current scope.
func (s *selectCoreAnalyzer) ExitSelectCore(node *tree.SelectCore) error {
	var err error
	s.detectedAttributes, err = s.context.relations()
	if err != nil {
		return err
	}

	s.oldScope()
	return nil
}

// EnterTableOrSubqueryTable adds the table to the list of used tables.
func (s *selectCoreAnalyzer) EnterTableOrSubqueryTable(node *tree.TableOrSubqueryTable) error {
	tbl, err := findTable(s.schemaTables, node.Name)
	if err != nil {
		return err
	}

	identifier := node.Name
	if node.Alias != "" {
		identifier = node.Alias
	}

	s.context.usedTables = append(s.context.usedTables, &types.Table{
		Name:        identifier,
		Columns:     tbl.Columns,
		Indexes:     tbl.Indexes,
		ForeignKeys: tbl.ForeignKeys,
	})

	return nil
}

// EnterResultColumnExpression adds the result column expression to the list of attributes
func (s *selectCoreAnalyzer) EnterResultColumnExpression(node *tree.ResultColumnExpression) error {
	s.context.addResult(node)
	return nil
}

// EnterResultColumnStar adds the result column expression to the list of attributes
func (s *selectCoreAnalyzer) EnterResultColumnStar(node *tree.ResultColumnStar) error {
	s.context.addResult(node)
	return nil
}

// EnterResultColumnTable adds the result column expression to the list of attributes
func (s *selectCoreAnalyzer) EnterResultColumnTable(node *tree.ResultColumnTable) error {
	s.context.addResult(node)
	return nil
}

// findTable finds a table by name
func findTable(tables []*types.Table, name string) (*types.Table, error) {
	for _, t := range tables {
		if t.Name == name {
			return t, nil
		}
	}

	return nil, fmt.Errorf(`table "%s" not found`, name)
}

// findColumn finds a column by name
func findColumn(columns []*types.Column, name string) (*types.Column, error) {
	for _, c := range columns {
		if c.Name == name {
			return c, nil
		}
	}

	return nil, fmt.Errorf(`column "%s" not found`, name)
}

// addTableIfNotPresent adds the table name to the column if it is not already present.
func addTableIfNotPresent(tableName string, expr tree.Walker) error {
	return expr.Walk(&tree.ImplementedWalker{
		FuncEnterExpressionColumn: func(col *tree.ExpressionColumn) error {
			if col.Table == "" {
				col.Table = tableName
			}
			return nil
		},
	})
}
