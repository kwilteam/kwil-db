package schema

import (
	"strings"
)

func buildCreateTable(name string, t Table[PGType, PGConstraint]) string {
	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(name)
	b.WriteString(" (")
	var i uint8
	for name, c := range t.Columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(name)
		b.WriteString(" ")
		b.WriteString(c.Type.String())
		for _, constraint := range c.Constraints {
			b.WriteString(" ")
			b.WriteString(constraint.String())
		}
		i++
	}
	b.WriteString(");")
	return b.String()
}

func buildCreateIndex(name string, i Index[PGIndex]) string {
	var b strings.Builder
	b.WriteString("CREATE INDEX ")
	b.WriteString(name)
	b.WriteString(" ON ")
	b.WriteString(i.Table)
	b.WriteString(" (")
	b.WriteString(i.Column)
	b.WriteString(");")
	return b.String()
}

func BuildDDL(db *Database[PGType, PGConstraint, PGIndex]) string {
	var sb strings.Builder
	sb.WriteString("BEGIN:\n")
	for name, t := range db.Tables {
		sb.WriteString(buildCreateTable(name, t))
		sb.WriteString("\n")
	}
	for name, i := range db.Indexes {
		sb.WriteString(buildCreateIndex(name, i))
		sb.WriteString("\n")
	}
	sb.WriteString("COMMIT;")
	return sb.String()
}
