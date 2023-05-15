package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/tree/sql-writer"
)

type Select struct {
	CTE        []*CTE
	SelectStmt *SelectStmt
}

func (s *Select) ToSQL() (str string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
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
	SelectCore *SelectCore
	OrderBy    *OrderBy
	Limit      *Limit
}

func (s *SelectStmt) ToSQL() (res string) {

	stmt := sqlwriter.NewWriter()
	stmt.WriteString(s.SelectCore.ToSQL())
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
	Columns    []string
	From       *FromClause
	Where      Expression
	GroupBy    *GroupBy
	Compound   *CompoundOperator
}

func (s *SelectCore) ToSQL() string {
	if s.From == nil {
		panic("from clause is required")
	}

	stmt := sqlwriter.NewWriter()
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
			stmt.WriteIdent(col)
		}
	}

	stmt.WriteString(s.From.ToSQL())
	if s.Where != nil {
		stmt.Token.Where()
		stmt.WriteString(s.Where.ToSQL())
	}
	if s.GroupBy != nil {
		stmt.WriteString(s.GroupBy.ToSQL())
	}
	if s.Compound != nil {
		stmt.WriteString(s.Compound.ToSQL())
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
	Operator     CompoundOperatorType
	SelectClause *SelectCore
}

func (c *CompoundOperator) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteString(c.Operator.ToSQL())
	stmt.WriteString(c.SelectClause.ToSQL())
	return stmt.String()
}
