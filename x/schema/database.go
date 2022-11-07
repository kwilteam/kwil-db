package schema

type Database struct {
	Name    string
	Tables  []Table
	Enums   []Enum
	Queries []Query
	Roles   []Role
}

type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name     string
	Type     string
	Nullable bool
}

type Enum struct {
	Name   string
	Values []string
}

type Role struct {
	Name    string
	Queries []string
}

type Query struct {
	Name      string
	Statement string
}
