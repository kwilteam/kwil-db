package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"
)

type Select struct {
	CTE        []*CTE
	SelectStmt *SelectStmt
}

func (s *Select) Accept(w Walker) error {
	return run(
		w.EnterSelect(s),
		acceptMany(w, s.CTE),
		accept(w, s.SelectStmt),
		w.ExitSelect(s),
	)
}

func (s *Select) ToSQL() (str string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err2, ok := r.(error)
			if !ok {
				err2 = fmt.Errorf("%v", r)
			}

			err = err2
		}
	}()

	stmt := sqlwriter.NewWriter()

	if len(s.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(s.CTE), func(i int) {
			stmt.WriteString(s.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(s.SelectStmt.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String(), nil
}

type SelectStmt struct {
	SelectCores []*SelectCore
	OrderBy     *OrderBy
	Limit       *Limit
}

func (s *SelectStmt) Accept(w Walker) error {
	return run(
		w.EnterSelectStmt(s),
		acceptMany(w, s.SelectCores),
		accept(w, s.OrderBy),
		accept(w, s.Limit),
		w.ExitSelectStmt(s),
	)
}

func (s *SelectStmt) ToSQL() (res string) {
	stmt := sqlwriter.NewWriter()
	for _, core := range s.SelectCores {
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

type SelectCore struct {
	SelectType SelectType
	Columns    []ResultColumn
	From       *FromClause
	Where      Expression
	GroupBy    *GroupBy
	Compound   *CompoundOperator
}

func (s *SelectCore) Accept(w Walker) error {
	return run(
		w.EnterSelectCore(s),
		acceptMany(w, s.Columns),
		accept(w, s.From),
		accept(w, s.Where),
		accept(w, s.GroupBy),
		accept(w, s.Compound),
		w.ExitSelectCore(s),
	)
}

func (s *SelectCore) ToSQL() string {
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

type FromClause struct {
	JoinClause *JoinClause
}

func (f *FromClause) Accept(w Walker) error {
	return run(
		w.EnterFromClause(f),
		accept(w, f.JoinClause),
		w.ExitFromClause(f),
	)
}

func (f *FromClause) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.Token.From()
	stmt.WriteString(f.JoinClause.ToSQL())
	return stmt.String()
}

type CompoundOperatorType uint8

const (
	CompoundOperatorTypeUnion CompoundOperatorType = iota
	CompoundOperatorTypeUnionAll
	CompoundOperatorTypeIntersect
	CompoundOperatorTypeExcept
)

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
	Operator CompoundOperatorType
}

func (c *CompoundOperator) Accept(w Walker) error {
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
