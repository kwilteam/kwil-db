package postgres

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"kwil/x/schemadef/hcl"
	"kwil/x/schemadef/hclspec"
	"kwil/x/schemadef/schema"

	"github.com/hashicorp/hcl/v2/hclparse"
)

func makeByte(s string) byte {
	var b byte
	if s == "" {
		return b
	}
	return []byte(s)[0]
}

func makeUint32Slice(in []uint64) []uint32 {
	out := make([]uint32, len(in))
	for i, v := range in {
		out[i] = uint32(v)
	}
	return out
}

func makeString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toPointer(x int) *int {
	return &x
}

func tname(t *schema.Table) (string, string) {
	if t.Schema != nil {
		return t.Schema.Name, t.Name
	}
	return "", t.Name
}

// ParseSchemaFiles parses the HCL files in the given paths. If a path represents a directory,
// its direct descendants will be considered, skipping any subdirectories. If a project file
// is present in the input paths, an error is returned.
func ParseSchemaFiles(paths ...string) (*schema.Realm, error) {
	p := hclparse.NewParser()
	for _, path := range paths {
		switch stat, err := os.Stat(path); {
		case err != nil:
			return nil, err
		case stat.IsDir():
			dir, err := os.ReadDir(path)
			if err != nil {
				return nil, err
			}
			for _, f := range dir {
				if f.IsDir() {
					continue
				}
				if err := mayParse(p, filepath.Join(path, f.Name())); err != nil {
					return nil, err
				}
			}
		default:
			if err := mayParse(p, path); err != nil {
				return nil, err
			}
		}
	}
	if len(p.Files()) == 0 {
		return nil, fmt.Errorf("no schema files found in: %s", paths)
	}

	var s schema.Realm
	if err := EvalHCL(p, &s, nil); err != nil {
		return nil, err
	}
	return &s, nil
}

// mayParse will parse the file in path if it is an HCL file.
func mayParse(p *hclparse.Parser, path string) error {
	if n := filepath.Base(path); filepath.Ext(n) != ".hcl" {
		return nil
	}
	switch _, diag := p.ParseHCLFile(path); {
	case diag.HasErrors():
		return diag
	default:
		return nil
	}
}

func IsComparisonOperator(s string) bool {
	switch s {
	case ">":
	case "<":
	case "<=":
	case ">=":
	case "=":
	case "<>":
	case "!=":
	default:
		return false
	}
	return true
}

func IsMathematicalOperator(s string) bool {
	switch s {
	case "+":
	case "-":
	case "*":
	case "/":
	case "%":
	case "^":
	case "|/":
	case "||/":
	case "!":
	case "!!":
	case "@":
	case "&":
	case "|":
	case "#":
	case "~":
	case "<<":
	case ">>":
	default:
		return false
	}
	return true
}

func precisionTypeAttr() *hcl.TypeAttr {
	return &hcl.TypeAttr{
		Name:     "precision",
		Kind:     reflect.Int,
		Required: false,
	}
}

func attr(typ *hcl.Type, key string) (*hcl.Attr, bool) {
	for _, a := range typ.Attrs {
		if a.K == key {
			return a, true
		}
	}
	return nil, false
}

func typeSpec(t schema.Type) (*hcl.Type, error) {
	if t, ok := t.(*schema.TimeType); ok && t.T != TypeDate {
		s := &hcl.Type{T: timeAlias(t.T)}
		if p := t.Precision; p != nil && *p != defaultTimePrecision {
			s.Attrs = []*hcl.Attr{hclspec.IntAttr("precision", *p)}
		}
		return s, nil
	}
	s, err := FormatType(t)
	if err != nil {
		return nil, err
	}
	return &hcl.Type{T: s}, nil
}

// formatTime overrides the default printing logic done by hcl.hclType.
func formatTime() hcl.TypeSpecOption {
	return hcl.WithTypeFormatter(func(t *hcl.Type) (string, error) {
		a, ok := attr(t, "precision")
		if !ok {
			return t.T, nil
		}
		p, err := a.Int()
		if err != nil {
			return "", fmt.Errorf(`postgres: parsing attribute "precision": %w`, err)
		}
		return FormatType(&schema.TimeType{T: t.T, Precision: &p})
	})
}

// generatedType returns the default and only type for a generated column.
func generatedType(string) string { return "STORED" }

func columnNames(cols []*schema.Column) []string {
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = c.Name
	}
	return names
}
