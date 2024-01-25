package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type UpsertType uint8

const (
	UpsertTypeDoNothing UpsertType = iota
	UpsertTypeDoUpdate
)

func (u UpsertType) Valid() error {
	switch u {
	case UpsertTypeDoNothing, UpsertTypeDoUpdate:
		return nil
	default:
		return fmt.Errorf("invalid upsert type: %d", u)
	}
}

type Upsert struct {
	node

	ConflictTarget *ConflictTarget
	Type           UpsertType
	Updates        []*UpdateSetClause
	Where          Expression
}

func (u *Upsert) Accept(v AstVisitor) any {
	return v.VisitUpsert(u)
}

func (u *Upsert) Walk(w AstWalker) error {
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
