package metadata

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
	Name  string
	Type  string
	Arity TypeArity
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
	Name    string
	Inputs  []Param
	Outputs []Param
}

type Param struct {
	Name  string
	Type  string
	Arity TypeArity
}

const (
	ScalarNumber   = "number"
	ScalarString   = "string"
	ScalarBool     = "bool"
	ScalarDate     = "date"
	ScalarTime     = "time"
	ScalarDateTime = "datetime"
	ScalarBytes    = "bytes"
	Void           = ""
)

type TypeArity int

const (
	Optional TypeArity = iota
	Required
	Repeated
)
