package schema

type Database struct {
	Owner       string           `yaml:"owner"`
	Name        string           `yaml:"name"`
	DefaultRole string           `yaml:"default_role"`
	Tables      map[string]Table `yaml:"tables"`
	Roles       map[string]Role  `yaml:"roles"`
	Queries     map[string]Query `yaml:"queries"`
	Indexes     map[string]Index `yaml:"indexes"`
}

type Table struct {
	Columns map[string]KuniformColumn `yaml:"columns"`
}

type Index struct {
	Table  string        `yaml:"table"`
	Column string        `yaml:"column"`
	Using  KuniformIndex `yaml:"using"`
}

type Role struct {
	Queries []string `yaml:"queries"`
}

type Query struct {
}
