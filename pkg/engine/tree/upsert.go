package tree

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type UpsertType uint8

const (
	UpsertTypeDoNothing UpsertType = iota
	UpsertTypeDoUpdate
)

type Upsert struct {
	ConflictTargetColumn string
	Type                 UpsertType
	Set                  map[string]Expression
	Where                *WhereClause
}

func (u *Upsert) toGoqu() exp.ConflictExpression {
	switch u.Type {
	case UpsertTypeDoNothing:
		return goqu.DoNothing()
	case UpsertTypeDoUpdate:
		ups := goqu.DoUpdate(u.ConflictTargetColumn, u.getSetRecords())
		if u.Where != nil {
			ups = ups.Where(u.Where.toGoquExpr())
		}
		return ups
	default:
		panic("invalid upsert type: " + string(u.Type))
	}
}

func (u *Upsert) getSetRecords() goqu.Record {
	setRecords := make(goqu.Record)
	for column, expression := range u.Set {
		setRecords[column] = expression.ToSqlStruct()
	}
	return setRecords
}
