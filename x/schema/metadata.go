package schema

type RequestMetadata struct {
	Wallet   string
	Database string
}

type Metadata struct {
	DbName      string
	Tables      []Table
	Enums       []Enum
	Queries     []Query
	Roles       []Role
	DefaultRole string
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
