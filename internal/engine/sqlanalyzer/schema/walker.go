package schema

import "github.com/kwilteam/kwil-db/parse/sql/tree"

// SchemaWalker walks statements and ensures that their statements are targeting a postgres schema / namespace.
type SchemaWalker struct {
	tree.AstListener
	schema string
	ctes   map[string]struct{} // we keep track of ctes since they should not be prefixed with a schema

	// SetCount is the number of table refs where the schema was set by the walker.
	SetCount int
}

func NewSchemaWalker(targetSchema string) *SchemaWalker {
	return &SchemaWalker{
		AstListener: tree.NewBaseListener(),
		schema:      targetSchema,
		ctes:        make(map[string]struct{}),
	}
}

type settable interface {
	SetSchema(string)
}

// set conditionally sets the schema of the settable, if the table is not a CTE.
func (s *SchemaWalker) set(table string, st settable) {
	if _, ok := s.ctes[table]; ok {
		return
	}

	s.SetCount++
	st.SetSchema(s.schema)
}

func (w *SchemaWalker) EnterInsertStmt(stmt *tree.InsertStmt) error {
	w.set(stmt.Table, stmt)
	return nil
}

func (w *SchemaWalker) EnterQualifiedTableName(q *tree.QualifiedTableName) error {
	w.set(q.TableName, q)
	return nil
}

func (w *SchemaWalker) EnterTableOrSubqueryTable(t *tree.TableOrSubqueryTable) error {
	w.set(t.Name, t)
	return nil
}

/*
	There is a special case where common table expressions should not be prefixed with a schema.
	In postgres, CTEs cannot be used before they are declared. For example, the following schema is will fail:
		with users_2 as (
			select * from users_1
		), users_1 as (
			select * from users
		)
		select * from users_2

	The following schema will succeed:
		with users_2 as (
			select * from users
		), users_1 as (
			select * from users_2
		)
		select * from users_1

	Therefore, we simply need to intercept the CTEs and keep track of them.
*/

func (w *SchemaWalker) EnterCTE(cte *tree.CTE) error {
	w.ctes[cte.Table] = struct{}{}
	return nil
}
