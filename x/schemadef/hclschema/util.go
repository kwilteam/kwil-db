package hclschema

import (
	"fmt"
	"strconv"

	"kwil/x/schemadef/hcl"
	"kwil/x/schemadef/sqlschema"

	"github.com/hashicorp/hcl/v2/hclparse"
)

// StrAttr is a helper method for constructing *hcl.Attr of type string.
func StrAttr(k, v string) *hcl.Attr {
	return &hcl.Attr{
		K: k,
		V: &hcl.LiteralValue{V: strconv.Quote(v)},
	}
}

// BoolAttr is a helper method for constructing *hcl.Attr of type bool.
func BoolAttr(k string, v bool) *hcl.Attr {
	return &hcl.Attr{
		K: k,
		V: &hcl.LiteralValue{V: strconv.FormatBool(v)},
	}
}

// IntAttr is a helper method for constructing *hcl.Attr with the numeric value of v.
func IntAttr(k string, v int) *hcl.Attr {
	return Int64Attr(k, int64(v))
}

// Int64Attr is a helper method for constructing *hcl.Attr with the numeric value of v.
func Int64Attr(k string, v int64) *hcl.Attr {
	return &hcl.Attr{
		K: k,
		V: &hcl.LiteralValue{V: strconv.FormatInt(v, 10)},
	}
}

// LitAttr is a helper method for constructing *hcl.Attr instances that contain literal values.
func LitAttr(k, v string) *hcl.Attr {
	return &hcl.Attr{
		K: k,
		V: &hcl.LiteralValue{V: v},
	}
}

// RawAttr is a helper method for constructing *hcl.Attr instances that contain sql expressions.
func RawAttr(k, v string) *hcl.Attr {
	return &hcl.Attr{
		K: k,
		V: &hcl.RawExpr{X: v},
	}
}

// VarAttr is a helper method for constructing *hcl.Attr instances that contain a variable reference.
func VarAttr(k, v string) *hcl.Attr {
	return &hcl.Attr{
		K: k,
		V: &hcl.Ref{V: v},
	}
}

// RefAttr is a helper method for constructing *hcl.Attr instances that contain a reference.
func RefAttr(k string, r *hcl.Ref) *hcl.Attr {
	return &hcl.Attr{
		K: k,
		V: r,
	}
}

// ListAttr is a helper method for constructing *hcl.Attr instances that contain list values.
func ListAttr(k string, litValues ...string) *hcl.Attr {
	lv := &hcl.ListValue{}
	for _, v := range litValues {
		lv.V = append(lv.V, &hcl.LiteralValue{V: v})
	}
	return &hcl.Attr{
		K: k,
		V: lv,
	}
}

// QualifyDuplicates sets the Qualified field equal to the schema name in any tables
// with duplicate names in the provided table specs.
func QualifyDuplicates(tableSpecs []*Table) error {
	seen := make(map[string]*Table, len(tableSpecs))
	for _, tbl := range tableSpecs {
		if s, ok := seen[tbl.Name]; ok {
			schemaName, err := SchemaName(s.Schema)
			if err != nil {
				return err
			}
			s.Qualifier = schemaName
			schemaName, err = SchemaName(tbl.Schema)
			if err != nil {
				return err
			}
			tbl.Qualifier = schemaName
		}
		seen[tbl.Name] = tbl
	}
	return nil
}

// QualifyReferences qualifies any reference with qualifier.
func QualifyReferences(tableSpecs []*Table, realm *sqlschema.Realm) error {
	type cref struct{ s, t string }
	byRef := make(map[cref]*Table)
	for _, t := range tableSpecs {
		r := cref{s: t.Qualifier, t: t.Name}
		if byRef[r] != nil {
			return fmt.Errorf("duplicate references were found for: %v", r)
		}
		byRef[r] = t
	}
	for _, t := range tableSpecs {
		sname, err := SchemaName(t.Schema)
		if err != nil {
			return err
		}
		s1, ok := realm.Schema(sname)
		if !ok {
			return fmt.Errorf("schema %q was not found in realm", sname)
		}
		t1, ok := s1.Table(t.Name)
		if !ok {
			return fmt.Errorf("table %q.%q was not found in realm", sname, t.Name)
		}
		for _, fk := range t.ForeignKeys {
			fk1, ok := t1.ForeignKey(fk.Symbol)
			if !ok {
				return fmt.Errorf("table %q.%q.%q was not found in realm", sname, t.Name, fk.Symbol)
			}
			for i, c := range fk.RefColumns {
				if r, ok := byRef[cref{s: fk1.RefTable.Schema.Name, t: fk1.RefTable.Name}]; ok && r.Qualifier != "" {
					fk.RefColumns[i] = qualifiedExternalColRef(fk1.RefColumns[i].Name, r.Name, r.Qualifier)
				} else if r, ok := byRef[cref{t: fk1.RefTable.Name}]; ok && r.Qualifier == "" {
					fk.RefColumns[i] = externalColRef(fk1.RefColumns[i].Name, r.Name)
				} else {
					return fmt.Errorf("missing reference for column %q in %q.%q.%q", c.V, sname, t.Name, fk.Symbol)
				}
			}
		}
	}
	return nil
}

// HCLBytesFunc returns a helper that evaluates an HCL document from a byte slice instead
// of from an hclparse.Parser instance.
func HCLBytesFunc(ev hcl.Evaluator) func(b []byte, v any, inp map[string]string) error {
	return func(b []byte, v any, inp map[string]string) error {
		parser := hclparse.NewParser()
		if _, diag := parser.ParseHCL(b, ""); diag.HasErrors() {
			return diag
		}
		return ev.Eval(parser, v, inp)
	}
}
