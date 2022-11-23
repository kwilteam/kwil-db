package schema

import (
	"fmt"
	"strings"
)

func buildCreateTable(name string, t Table) ([]string, error) {
	var bs []string
	var alters []string
	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString(name)
	b.WriteString(" (")
	var i uint8
	for colname, c := range t.Columns {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(colname)
		b.WriteString(" ")
		b.WriteString(Types[c.Type].String())
		i++

		alterStmts, err := c.BuildAttributes(name, colname)
		if err != nil {
			return nil, err
		}
		alters = append(alters, alterStmts...)
	}
	b.WriteString(");\n")
	bs = append(bs, b.String())
	bs = append(bs, alters...)
	return bs, nil
}

func (c *KuniformColumn) BuildAttributes(tableName, columnName string) ([]string, error) {
	var stmts []string
	atts, err := c.GetAttributes()
	if err != nil {
		return nil, err
	}

	if len(atts) == 0 {
		return stmts, nil
	}

	for _, att := range atts {
		switch att { // can include MinLength, MaxLength, PrimaryKey, Unique, NotNull, Default
		case KuniformMinLength:
			if c.Type != KuniformString {
				return stmts, fmt.Errorf("cannot set MinLength on non-string column %s", columnName)
			}
			var b strings.Builder
			b.WriteString("ALTER TABLE ")
			b.WriteString(tableName)
			b.WriteString(" ADD CONSTRAINT ")
			b.WriteString(tableName)
			b.WriteString("_")
			b.WriteString(columnName)
			b.WriteString("_minlength CHECK (length(")
			b.WriteString(columnName)
			b.WriteString(") >= ")
			b.WriteString(c.Attributes[KuniformMinLength])
			b.WriteString(");\n")
			stmts = append(stmts, b.String())
		case KuniformMaxLength:
			if c.Type != KuniformString {
				return stmts, fmt.Errorf("cannot set MaxLength on non-string column %s", columnName)
			}
			var b strings.Builder
			b.WriteString("ALTER TABLE ")
			b.WriteString(tableName)
			b.WriteString(" ADD CONSTRAINT ")
			b.WriteString(tableName)
			b.WriteString("_")
			b.WriteString(columnName)
			b.WriteString("_maxlength CHECK (length(")
			b.WriteString(columnName)
			b.WriteString(") <= ")
			b.WriteString(c.Attributes[KuniformMaxLength])
			b.WriteString(");\n")
			stmts = append(stmts, b.String())
		case KuniformPrimaryKey:
			var b strings.Builder
			b.WriteString("ALTER TABLE ")
			b.WriteString(tableName)
			b.WriteString(" ADD PRIMARY KEY (")
			b.WriteString(columnName)
			b.WriteString(");\n")
			stmts = append(stmts, b.String())
		case KuniformUnique:
			var b strings.Builder
			b.WriteString("ALTER TABLE ")
			b.WriteString(tableName)
			b.WriteString(" ADD CONSTRAINT ")
			b.WriteString(tableName)
			b.WriteString("_")
			b.WriteString(columnName)
			b.WriteString("_unique UNIQUE (")
			b.WriteString(columnName)
			b.WriteString(");\n")
			stmts = append(stmts, b.String())
		case KuniformNotNull:
			var b strings.Builder
			b.WriteString("ALTER TABLE ")
			b.WriteString(tableName)
			b.WriteString(" ALTER COLUMN ")
			b.WriteString(columnName)
			b.WriteString(" SET NOT NULL;")
			stmts = append(stmts, b.String())
		case KuniformDefault:
			var b strings.Builder
			b.WriteString("ALTER TABLE ")
			b.WriteString(tableName)
			b.WriteString(" ALTER COLUMN ")
			b.WriteString(columnName)
			b.WriteString(" SET DEFAULT ")
			b.WriteString(fmt.Sprintf("%v", c.Attributes[KuniformDefault]))
			b.WriteString(";\n")
			stmts = append(stmts, b.String())
		default:
			return nil, fmt.Errorf("unknown attribute %s", att)
		}
	}
	return stmts, nil
}

func buildCreateIndex(name string, i Index) string {
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

func (db *Database) GenerateDDL() (string, error) {
	var sb strings.Builder
	sb.WriteString("BEGIN:\n")
	for name, t := range db.Tables {
		stmts, err := buildCreateTable(name, t)
		if err != nil {
			return "", err
		}

		for _, s := range stmts {
			sb.WriteString(s)
		}
	}
	for name, i := range db.Indexes {
		sb.WriteString(buildCreateIndex(name, i))
	}
	sb.WriteString("COMMIT;")
	return sb.String(), nil
}

func (c *KuniformColumn) GetAttributes() ([]KuniformAttribute, error) {
	var atts []KuniformAttribute
	for att := range c.Attributes {
		if Attributes[att] {
			atts = append(atts, att)
		} else {
			return nil, fmt.Errorf("unknown attribute %s", att)
		}
	}
	return atts, nil
}
