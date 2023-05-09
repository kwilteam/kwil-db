package tree

type UpsertType uint8

const (
	UpsertTypeDoNothing UpsertType = iota
	UpsertTypeDoUpdate
)

type Upsert struct {
	stmt           *upsertBuilder
	ConflictTarget *ConflictTarget
	Type           UpsertType
	Updates        []*UpdateSetClause
	Where          Expression
}

func (u *Upsert) ToSQL() string {
	u.stmt = Builder.BeginUpsert()
	if u.ConflictTarget != nil {
		u.stmt.ConflictTarget(u.ConflictTarget)
	}

	switch u.Type {
	case UpsertTypeDoNothing:
		u.stmt.DoNothing()
	case UpsertTypeDoUpdate:
		u.stmt.DoUpdate(u.Updates)
		if u.Where != nil {
			u.stmt.Where(u.Where)
		}
	}

	return u.stmt.String()
}

type upsertBuilder struct {
	stmt *sqlBuilder
}

func (b *builder) BeginUpsert() *upsertBuilder {
	u := &upsertBuilder{
		stmt: newSQLBuilder(),
	}
	u.stmt.Write(SPACE, ON, SPACE, CONFLICT, SPACE)
	return u
}

func (b *upsertBuilder) ConflictTarget(ct *ConflictTarget) {
	b.stmt.WriteString(ct.ToSQL())
}

func (b *upsertBuilder) String() string {
	return b.stmt.String()
}

func (b *upsertBuilder) DoNothing() {
	b.stmt.Write(SPACE, DO, SPACE, NOTHING, SPACE)
}

func (b *upsertBuilder) DoUpdate(setClause []*UpdateSetClause) {
	b.stmt.Write(SPACE, DO, SPACE, UPDATE, SPACE, SET, SPACE)
	for i, set := range setClause {
		if i > 0 && i < len(setClause) {
			b.stmt.Write(COMMA, SPACE)
		}
		b.stmt.WriteString(set.ToSQL())
	}
	b.stmt.Write(SPACE)
}

func (b *upsertBuilder) Where(expression Expression) {
	b.stmt.Write(SPACE, WHERE, SPACE)
	b.stmt.WriteString(expression.ToSQL())
}
