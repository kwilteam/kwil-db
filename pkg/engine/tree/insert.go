package tree

import (
	"fmt"
)

type Insert struct {
	stmt            *insertBuilder
	InsertType      InsertType
	Table           string
	TableAlias      string
	Columns         []string
	Values          [][]InsertExpression
	Upsert          *Upsert
	ReturningClause *ReturningClause
}

type InsertType uint8

const (
	InsertTypeInsert InsertType = iota
	InsertTypeReplace
	InsertTypeInsertOrReplace
)

func (i *Insert) ToSql() (result string, err error) {
	defer func() {
		if err == nil {
			return
		}
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in InsertStatement.ToSql: %v", r)
		}
	}()

	if i.Table == "" {
		return "", fmt.Errorf("sql syntax error: insert does not contain table name")
	}

	i.stmt = Builder.BeginInsert()
	i.stmt.Insert(i.InsertType)
	i.stmt.Table(i.Table)
	if i.TableAlias != "" {
		i.stmt.TableAlias(i.TableAlias)
	}
	if len(i.Columns) > 0 {
		i.stmt.Columns(i.Columns...)
	}

	if len(i.Values) > 0 {
		i.stmt.Values(i.Values...)
	}

	if i.Upsert != nil {
		i.stmt.Upsert(i.Upsert)
	}

	if i.ReturningClause != nil {
		i.stmt.Returning(i.ReturningClause)
	}

	return i.stmt.ToSQL(), nil
}

func (b *builder) BeginInsert() *insertBuilder {
	return &insertBuilder{
		stmt: newSQLBuilder(),
	}
}

type insertBuilder struct {
	stmt *sqlBuilder
}

func (b *insertBuilder) Insert(insertType InsertType) {
	switch insertType {
	case InsertTypeInsert:
		b.stmt.Write(SPACE, INSERT, SPACE, INTO, SPACE)
	case InsertTypeReplace:
		b.stmt.Write(SPACE, REPLACE, SPACE, INTO, SPACE)
	case InsertTypeInsertOrReplace:
		b.stmt.Write(SPACE, INSERT, SPACE, OR, SPACE, REPLACE, SPACE, INTO, SPACE)
	}
}

func (b *insertBuilder) Table(tbl string) {
	b.stmt.Write(SPACE)
	b.stmt.WriteIdent(tbl)
	b.stmt.Write(SPACE)
}

func (b *insertBuilder) TableAlias(alias string) {
	b.stmt.Write(SPACE, AS, SPACE)
	b.stmt.WriteIdent(alias)
	b.stmt.Write(SPACE)
}

func (b *insertBuilder) Columns(columns ...string) {
	b.stmt.Write(SPACE, LPAREN)
	for i, col := range columns {
		if i > 0 && i < len(columns) {
			b.stmt.Write(COMMA, SPACE)
		}
		b.stmt.WriteIdent(col)
	}
	b.stmt.Write(RPAREN)
}

func (b *insertBuilder) Values(values ...[]InsertExpression) {
	b.stmt.Write(SPACE, VALUES, SPACE)
	for i, value := range values {
		if i > 0 && i < len(values) {
			b.stmt.Write(COMMA, SPACE)
		}
		b.singleValues(value...)
	}
}

func (b *insertBuilder) singleValues(values ...InsertExpression) {
	b.stmt.Write(LPAREN)
	for i, value := range values {
		if i > 0 && i < len(values) {
			b.stmt.Write(COMMA, SPACE)
		}
		b.stmt.WriteString(value.ToSQL())
	}
	b.stmt.Write(RPAREN)
}

func (b *insertBuilder) Upsert(upsert *Upsert) {
	b.stmt.Write(SPACE)
	b.stmt.WriteString(upsert.ToSQL())
}

func (b *insertBuilder) Returning(returning *ReturningClause) {
	b.stmt.Write(SPACE)
	b.stmt.WriteString(returning.ToSQL())
}

func (b *insertBuilder) ToSQL() string {
	b.stmt.Write(SEMICOLON)
	return b.stmt.String()
}
