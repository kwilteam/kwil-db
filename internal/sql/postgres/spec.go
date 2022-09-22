package postgres

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/kwilteam/kwil-db/internal/hcl"
	"github.com/kwilteam/kwil-db/internal/schema"
	"github.com/kwilteam/kwil-db/internal/spec"
)

type (
	Document struct {
		Tables  []*spec.Table  `spec:"table"`
		Enums   []*Enum        `spec:"enum"`
		Schemas []*spec.Schema `spec:"schema"`
	}

	Enum struct {
		Name   string   `spec:",name"`
		Schema *hcl.Ref `spec:"schema"`
		Values []string `spec:"values"`
		hcl.DefaultExtension
	}
)

func init() {
	hcl.Register("enum", &Enum{})
}

var hclState = hcl.New(
	hcl.WithTypes(TypeRegistry.Specs()),
	hcl.WithScopedEnums("table.index.type", IndexTypeBTree, IndexTypeHash, IndexTypeGIN, IndexTypeGiST, IndexTypeBRIN),
	hcl.WithScopedEnums("table.column.as.type", "STORED"),
	hcl.WithScopedEnums("table.foreign_key.on_update", spec.ReferenceVars...),
	hcl.WithScopedEnums("table.foreign_key.on_delete", spec.ReferenceVars...),
)

func ParsePaths(paths ...string) (*schema.Realm, error) {
	var d Document
	parser, err := hcl.ParsePaths(paths...)
	if err != nil {
		return nil, err
	}
	if err := hclState.Eval(parser, &d, nil); err != nil {
		return nil, err
	}

	r := &schema.Realm{}
	if err := spec.Scan(r, d.Schemas, d.Tables, convertTable); err != nil {
		return nil, fmt.Errorf("spec: failed converting to *schema.Realm: %w", err)
	}
	if len(d.Enums) > 0 {
		if err := convertEnums(d.Tables, d.Enums, r); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func ParseBytes(b []byte) (*schema.Realm, error) {
	var d Document
	parser := hclparse.NewParser()
	if _, diag := parser.ParseHCL(b, ""); diag.HasErrors() {
		return nil, diag
	}

	if err := hclState.Eval(parser, &d, nil); err != nil {
		return nil, err
	}

	r := &schema.Realm{}
	if err := spec.Scan(r, d.Schemas, d.Tables, convertTable); err != nil {
		return nil, fmt.Errorf("spec: failed converting to *schema.Realm: %w", err)
	}
	if len(d.Enums) > 0 {
		if err := convertEnums(d.Tables, d.Enums, r); err != nil {
			return nil, err
		}
	}
	return r, nil
}
