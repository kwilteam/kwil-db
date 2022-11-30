package schema

type Database struct {
	Owner       string           `yaml:"owner"`
	Name        string           `yaml:"name"`
	DefaultRole string           `yaml:"default_role"`
	Tables      Tables           `yaml:"tables"`
	Roles       Roles            `yaml:"roles"`
	Queries     map[string]Query `yaml:"queries"`
	Indexes     Indices          `yaml:"indexes"`
}

type Tables map[string]Table

type Table struct {
	Columns map[string]KuniformColumn `yaml:"columns"`
}

type Indices map[string]Index

type Index struct {
	Table  string        `yaml:"table"`
	Column string        `yaml:"column"`
	Using  KuniformIndex `yaml:"using"`
}

type Roles map[string]Role

type Role struct {
	Queries []string `yaml:"queries"`
}

type Query struct {
}
