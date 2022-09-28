package hclspec

import "kwil/x/schemadef/hcl"

type (
	Document struct {
		Tables  []*Table  `spec:"table"`
		Enums   []*Enum   `spec:"enum"`
		Schemas []*Schema `spec:"schema"`
		Queries []*Query  `spec:"query"`
		Roles   []*Role   `spec:"role"`
	}

	// Schema holds a specification for a Schema.
	Schema struct {
		Name string `spec:"name,name"`
		hcl.DefaultExtension
	}

	// Table holds a specification for an SQL table.
	Table struct {
		Name        string        `spec:",name"`
		Qualifier   string        `spec:",qualifier"`
		Schema      *hcl.Ref      `spec:"schema"`
		Columns     []*Column     `spec:"column"`
		PrimaryKey  *PrimaryKey   `spec:"primary_key"`
		ForeignKeys []*ForeignKey `spec:"foreign_key"`
		Indexes     []*Index      `spec:"index"`
		Checks      []*Check      `spec:"check"`
		hcl.DefaultExtension
	}

	// Column holds a specification for a column in an SQL table.
	Column struct {
		Name     string    `spec:",name"`
		Nullable bool      `spec:"null"`
		Type     *hcl.Type `spec:"type"`
		Default  hcl.Value `spec:"default"`
		hcl.DefaultExtension
	}

	// PrimaryKey holds a specification for the primary key of a table.
	PrimaryKey struct {
		Columns []*hcl.Ref `spec:"columns"`
		hcl.DefaultExtension
	}

	// Index holds a specification for the index key of a table.
	Index struct {
		Name    string       `spec:",name"`
		Unique  bool         `spec:"unique,omitempty"`
		Parts   []*IndexPart `spec:"on"`
		Columns []*hcl.Ref   `spec:"columns"`
		hcl.DefaultExtension
	}

	// IndexPart holds a specification for the index key part.
	IndexPart struct {
		Desc   bool     `spec:"desc,omitempty"`
		Column *hcl.Ref `spec:"column"`
		Expr   string   `spec:"expr,omitempty"`
		hcl.DefaultExtension
	}

	// Check holds a specification for a check constraint on a table.
	Check struct {
		Name string `spec:",name"`
		Expr string `spec:"expr"`
		hcl.DefaultExtension
	}
	// ForeignKey holds a specification for the Foreign key of a table.
	ForeignKey struct {
		Symbol     string     `spec:",name"`
		Columns    []*hcl.Ref `spec:"columns"`
		RefColumns []*hcl.Ref `spec:"ref_columns"`
		OnUpdate   *hcl.Ref   `spec:"on_update"`
		OnDelete   *hcl.Ref   `spec:"on_delete"`
		hcl.DefaultExtension
	}

	Enum struct {
		Name   string   `spec:",name"`
		Schema *hcl.Ref `spec:"schema"`
		Values []string `spec:"values"`
		hcl.DefaultExtension
	}

	// Query holds a specification for a query.
	Query struct {
		Name      string `spec:",name"`
		Statement string `spec:"statement"`
		hcl.DefaultExtension
	}

	// Role holds a specification for a role.
	Role struct {
		Name    string     `spec:",name"`
		Queries []*hcl.Ref `spec:"queries"`
		Default bool       `spec:"default,omitempty"`
		hcl.DefaultExtension
	}

	// Type represents a database agnostic column type.
	Type string
)

func init() {
	hcl.Register("table", &Table{})
	hcl.Register("schema", &Schema{})
	hcl.Register("query", &Query{})
	hcl.Register("role", &Role{})
	hcl.Register("enum", &Enum{})
}
