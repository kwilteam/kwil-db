package schema

import (
	"fmt"
	"kwil/x/crypto"
	"strings"

	conv "github.com/cstockton/go-conv"
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
			b.WriteString(constraintName(tableName, columnName, "MinLength"))
			b.WriteString(" CHECK (length(")
			b.WriteString(columnName)
			b.WriteString(") >= ")
			kl := c.Attributes[KuniformMinLength]
			_, err := conv.Int32(kl)
			if err != nil {
				return stmts, fmt.Errorf("MinLength must be an int32. Received %v", kl)
			}
			b.WriteString(kl)
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
			b.WriteString(constraintName(tableName, columnName, "MaxLength"))
			b.WriteString(" CHECK (length(")
			b.WriteString(columnName)
			b.WriteString(") <= ")
			kl := c.Attributes[KuniformMaxLength]
			_, err := conv.Int32(kl)
			if err != nil {
				return stmts, fmt.Errorf("MaxLength must be an int32. Received %v", kl)
			}
			b.WriteString(kl)
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
			// TODO: dynamically determine the type of the default value
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

func constraintName(tableName, columnName, constraintType string) string {

	return constraintType[:1] + crypto.Sha224Str([]byte(tableName+"_"+columnName+"_"+constraintType))
}

func As[T any](into T, from any) error {
	return conv.Infer(into, from)
}

//func (c *KuniformColumn) ToColumnType(s string)

func buildCreateIndex(name string, i Index) string {
	var b strings.Builder
	indNm := crypto.Sha224Str([]byte(name))
	b.WriteString("CREATE INDEX ")
	b.WriteString(indNm)
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
		stmts, err := buildCreateTable(db.addSchema(name), t)
		if err != nil {
			return "", err
		}

		for _, s := range stmts {
			sb.WriteString(s)
		}
	}
	for name, i := range db.Indexes {
		sb.WriteString(buildCreateIndex(db.addSchema(name), i))
	}
	sb.WriteString("\nCOMMIT;")
	return sb.String(), nil
}

func (db *Database) addSchema(s string) string {
	return db.Owner + "_" + db.Name + "." + s
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
