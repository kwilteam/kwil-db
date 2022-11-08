package sqlspec

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"ksl/kslwrite"
	"ksl/sqlutil"
)

type Marshaler interface {
	MarshalSpec(*Realm) ([]byte, error)
}

var kw = kslwrite.Builder{}

func Marshal(w io.Writer, r *Realm) error {
	m := newSpecMarshaler()
	data, err := m.MarshalSpec(r)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func MarshalSpec(r *Realm) ([]byte, error) {
	m := newSpecMarshaler()
	return m.MarshalSpec(r)
}

type specMarshaler struct {
	file *kslwrite.File
}

func newSpecMarshaler() *specMarshaler {
	return &specMarshaler{&kslwrite.File{}}
}

func (m *specMarshaler) MarshalSpec(r *Realm) ([]byte, error) {
	for _, s := range r.Schemas {
		for _, t := range s.Tables {
			if err := m.marshalTable(t); err != nil {
				return nil, err
			}
		}
	}

	var buf bytes.Buffer
	err := kslwrite.Marshal(&buf, m.file)
	return buf.Bytes(), err
}

func (m *specMarshaler) marshalTable(t *Table) error {
	block := m.file.GetOrCreateBlock("table", t.Name)
	for _, c := range t.Columns {
		if err := m.marshalColumn(c); err != nil {
			return err
		}
	}

	for _, fk := range t.ForeignKeys {
		if err := m.marshalForeignKey(fk); err != nil {
			return err
		}
	}

	if t.PrimaryKey != nil {
		if len(t.PrimaryKey.Parts) == 1 {
			col := t.PrimaryKey.Parts[0].Column
			block.GetOrCreateDefinition(col.Name).AddAnnotations(kw.Annot("id"))
		} else {
			columns := make([]string, len(t.PrimaryKey.Parts))
			for i, p := range t.PrimaryKey.Parts {
				columns[i] = p.Column.Name
			}
			block.AddAnnotations(kw.Annot("id").AddArgs(kslwrite.List(columns...)))
		}
	}

	for _, idx := range t.Indexes {
		if len(idx.Parts) == 1 {
			col := idx.Parts[0].Column
			if idx.Unique && idx.Name == DefaultUniqueIndexName(t, col) {
				block.GetOrCreateDefinition(col.Name).AddAnnotations(kw.Annot("unique"))
			} else {
				block := m.file.GetOrCreateBlock("table", t.Name)

				annot := kw.Annot("index")
				if idx.Name != DefaultIndexName(t, col) {
					annot.AddArgs(idx.Name)
				}
				// if typePart := (IndexType{}); has(idx.Attrs, &typePart) {
				// 	annot.AddKwarg("type", typePart.T)
				// }
				if idx.Unique {
					annot.AddKwarg("unique", "true")
				}
				block.GetOrCreateDefinition(col.Name).AddAnnotations(annot)
			}
		} else {
			columns := make([]*Column, len(idx.Parts))
			for i, p := range idx.Parts {
				columns[i] = p.Column
			}
			columnNames := make([]string, len(idx.Parts))
			for i, p := range idx.Parts {
				columnNames[i] = p.Column.Name
			}

			annot := kw.Annot("index")
			if idx.Name != DefaultIndexName(t, columns...) {
				annot.AddArgs(idx.Name)
			}

			annot.AddKwarg("columns", kslwrite.List(columnNames...))
			if idx.Unique {
				annot.AddKwarg("unique", "true")
			}
			if typePart := (IndexType{}); has(idx.Attrs, &typePart) {
				annot.AddKwarg("type", typePart.T)
			}
			block.AddAnnotations(annot)
		}
	}
	return nil
}

func (m *specMarshaler) marshalForeignKey(fk *ForeignKey) error {
	switch {
	case len(fk.Columns) != len(fk.RefColumns):
		return fmt.Errorf("foreign key %s.%s has mismatched column and ref column counts", fk.Table.Name, fk.Name)
	case len(fk.Columns) == 0:
		return fmt.Errorf("foreign key %s.%s has no columns", fk.Table.Name, fk.Name)
	case len(fk.Columns) == 1:
		qual := newColumnQualifier(fk.Table.Schema.Name, fk.Table.Name)
		annot := kw.Annot("foreign_key").AddArgs(qual.Qualify(fk.RefColumns[0]))

		if fk.Name != DefaultForeignKeyName(fk.Table, fk.Columns[0]) {
			annot.AddKwarg("name", kslwrite.Quoted(fk.Name))
		}
		if fk.OnDelete != "" && fk.OnDelete != "NO ACTION" {
			annot.AddKwarg("on_delete", kslwrite.Quoted(fk.OnDelete))
		}
		if fk.OnUpdate != "" && fk.OnUpdate != "NO ACTION" {
			annot.AddKwarg("on_update", kslwrite.Quoted(fk.OnUpdate))
		}
		m.file.GetOrCreateBlock("table", fk.Table.Name).GetOrCreateDefinition(fk.Columns[0].Name).AddAnnotations(annot)
	default:
		qual := newColumnQualifier(fk.Table.Schema.Name, fk.Table.Name)

		annot := kw.Annot("foreign_key")
		if fk.Name != DefaultForeignKeyName(fk.Table, fk.Columns...) {
			annot.AddArgs(kslwrite.Quoted(fk.Name))
		}
		columns := make([]string, len(fk.Columns))
		refColumns := make([]string, len(fk.RefColumns))
		for i, c := range fk.Columns {
			columns[i] = qual.Qualify(c)
		}
		for i, c := range fk.RefColumns {
			refColumns[i] = qual.Qualify(c)
		}

		annot.AddKwarg("columns", kslwrite.List(columns...))
		annot.AddKwarg("references", kslwrite.List(refColumns...))

		if fk.OnDelete != "" && fk.OnDelete != "NO ACTION" {
			annot.AddKwarg("on_delete", kslwrite.Quoted(fk.OnDelete))
		}
		if fk.OnUpdate != "" && fk.OnUpdate != "NO ACTION" {
			annot.AddKwarg("on_update", kslwrite.Quoted(fk.OnUpdate))
		}
		m.file.GetOrCreateBlock("table", fk.Table.Name).AddAnnotations(annot)
	}
	return nil
}

func (m *specMarshaler) marshalColumn(c *Column) error {
	block := m.file.GetOrCreateBlock("table", c.Table.Name)
	def := block.GetOrCreateDefinition(c.Name)

	typ := c.Type.Raw
	if t, err := TypeRegistry.Convert(c.Type.Type); err == nil {
		typ = t.Type
		for _, a := range t.Attrs {
			switch a.Name {
			case "array":
				arr, _ := a.Bool()
				def.SetArray(arr)
			default:
				s, _ := a.String()
				def.AddAnnotations(kw.Annot(a.Name).AddArgs(s))
			}
		}
	}
	def.SetType(typ)

	if c.Default != nil {
		switch val := c.Default.(type) {
		case *LiteralExpr:
			v := val.Value
			if IsStringType(typ) {
				v = kslwrite.Quoted(v)
			}
			def.AddAnnotations(kw.Annot("default").AddArgs(v))
		case *RawExpr:
			def.AddAnnotations(kw.Annot("default").AddArgs(sqlutil.TrimCast(val.Expr)))
		}
	}

	if c.Type.Nullable {
		def.SetOptional(true)
	}
	return nil
}

type columnQualifier struct{ schemaName, tableName string }

func newColumnQualifier(schemaName, tableName string) *columnQualifier {
	return &columnQualifier{schemaName, tableName}
}

func (q *columnQualifier) Qualify(c *Column) string {
	if c.Table.Schema.Name == q.schemaName {
		if c.Table.Name == q.tableName {
			return c.Name
		}
		return c.Table.Name + "." + c.Name
	}
	return c.Table.Schema.Name + "." + c.Table.Name + "." + c.Name
}

func IsStringType(t string) bool {
	switch strings.ToLower(t) {
	case "string", "varchar", "text":
		return true
	}
	return false
}
