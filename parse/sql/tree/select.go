package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type SelectStmt struct {
	node

	CTE  []*CTE
	Stmt *SelectCore
}

func (s *SelectStmt) Accept(v AstVisitor) any {
	return v.VisitSelectStmt(s)
}

func (s *SelectStmt) Walk(w AstListener) error {
	return run(
		w.EnterSelectStmt(s),
		walkMany(w, s.CTE),
		walk(w, s.Stmt),
		w.ExitSelectStmt(s),
	)
}

func (s *SelectStmt) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(s.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(s.CTE), func(i int) {
			stmt.WriteString(s.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(s.Stmt.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String()
}

func (s *SelectStmt) statement() {}

type SelectCore struct {
	node

	SimpleSelects []*SimpleSelect
	OrderBy       *OrderBy
	Limit         *Limit
}

func (s *SelectCore) Accept(v AstVisitor) any {
	return v.VisitSelectCore(s)
}

func (s *SelectCore) Walk(w AstListener) error {
	return run(
		w.EnterSelectCore(s),
		walkMany(w, s.SimpleSelects),
		walk(w, s.OrderBy),
		walk(w, s.Limit),
		w.ExitSelectCore(s),
	)
}

func (s *SelectCore) ToSQL() (res string) {
	stmt := sqlwriter.NewWriter()
	for _, core := range s.SimpleSelects {
		stmt.WriteString(core.ToSQL())
	}
	if s.OrderBy != nil {
		stmt.WriteString(s.OrderBy.ToSQL())
	}
	if s.Limit != nil {
		stmt.WriteString(s.Limit.ToSQL())
	}

	return stmt.String()
}

type SimpleSelect struct {
	node

	SelectType SelectType
	Columns    []ResultColumn
	From       Relation
	Where      Expression
	GroupBy    *GroupBy
	Compound   *CompoundOperator
}

func (s *SimpleSelect) Accept(v AstVisitor) any {
	return v.VisitSimpleSelect(s)
}

func (s *SimpleSelect) Walk(w AstListener) error {
	return run(
		w.EnterSimpleSelect(s),
		walkMany(w, s.Columns),
		walk(w, s.From),
		walk(w, s.Where),
		walk(w, s.GroupBy),
		walk(w, s.Compound),
		w.ExitSimpleSelect(s),
	)
}

func (s *SimpleSelect) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if s.Compound != nil {
		stmt.WriteString(s.Compound.ToSQL())
	}

	stmt.Token.Select()
	if s.SelectType == SelectTypeDistinct {
		stmt.Token.Distinct()
	}

	if len(s.Columns) == 0 {
		stmt.Token.Asterisk()
	} else {
		for i, col := range s.Columns {
			if i > 0 && i < len(s.Columns) {
				stmt.Token.Comma()
			}
			stmt.WriteString(col.ToSQL())
		}
	}

	if s.From != nil {
		stmt.Token.From()
		stmt.WriteString(s.From.ToSQL())
	}
	if s.Where != nil {
		stmt.Token.Where()
		stmt.WriteString(s.Where.ToSQL())
	}
	if s.GroupBy != nil {
		stmt.WriteString(s.GroupBy.ToSQL())
	}
	return stmt.String()
}

type SelectType uint8

const (
	SelectTypeAll SelectType = iota
	SelectTypeDistinct
)

func (s SelectType) Valid() error {
	switch s {
	case SelectTypeAll, SelectTypeDistinct:
		return nil
	default:
		return fmt.Errorf("invalid select type: %d", s)
	}
}

type CompoundOperatorType uint8

const (
	CompoundOperatorTypeUnion CompoundOperatorType = iota
	CompoundOperatorTypeUnionAll
	CompoundOperatorTypeIntersect
	CompoundOperatorTypeExcept
)

func (c CompoundOperatorType) Valid() error {
	switch c {
	case CompoundOperatorTypeUnion, CompoundOperatorTypeUnionAll, CompoundOperatorTypeIntersect, CompoundOperatorTypeExcept:
		return nil
	default:
		return fmt.Errorf("invalid compound operator type: %d", c)
	}
}

func (c *CompoundOperatorType) ToSQL() string {
	switch *c {
	case CompoundOperatorTypeUnion:
		return "UNION"
	case CompoundOperatorTypeUnionAll:
		return "UNION ALL"
	case CompoundOperatorTypeIntersect:
		return "INTERSECT"
	case CompoundOperatorTypeExcept:
		return "EXCEPT"
	default:
		panic(fmt.Errorf("unknown compound operator type %d", *c))
	}
}

type CompoundOperator struct {
	node

	Operator CompoundOperatorType
}

func (c *CompoundOperator) Accept(v AstVisitor) any {
	return v.VisitCompoundOperator(c)
}

func (c *CompoundOperator) AcceptVisitor(v AstVisitor) any {
	return v.VisitCompoundOperator(c)
}

func (c *CompoundOperator) Walk(w AstListener) error {
	return run(
		w.EnterCompoundOperator(c),
		w.ExitCompoundOperator(c),
	)
}

func (c *CompoundOperator) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteString(c.Operator.ToSQL())
	return stmt.String()
}
