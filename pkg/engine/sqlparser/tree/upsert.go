package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"
)

type UpsertType uint8

const (
	UpsertTypeDoNothing UpsertType = iota
	UpsertTypeDoUpdate
)

type Upsert struct {
	ConflictTarget *ConflictTarget
	Type           UpsertType
	Updates        []*UpdateSetClause
	Where          Expression
}

func (u *Upsert) Accept(w Walker) error {
	return run(
		w.EnterUpsert(u),
		accept(w, u.ConflictTarget),
		acceptMany(w, u.Updates),
		accept(w, u.Where),
		w.ExitUpsert(u),
	)
}

func (u *Upsert) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	stmt.Token.On().Conflict()

	if u.ConflictTarget != nil {
		stmt.WriteString(u.ConflictTarget.ToSQL())
	}

	switch u.Type {
	case UpsertTypeDoNothing:
		stmt.Token.Do().Nothing()
	case UpsertTypeDoUpdate:
		stmt.Token.Do().Update().Set()

		stmt.WriteList(len(u.Updates), func(i int) {
			stmt.WriteString(u.Updates[i].ToSQL())
		})

		if u.Where != nil {
			stmt.Token.Where()
			stmt.WriteString(u.Where.ToSQL())
		}
	}

	return stmt.String()
}
