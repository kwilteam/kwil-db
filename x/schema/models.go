package schema

type Database[T GType, C GConstraint, I GIndex] struct {
	Owner       string                 `yaml:"owner"`
	Name        string                 `yaml:"name"`
	DefaultRole string                 `yaml:"default_role"`
	Tables      map[string]Table[T, C] `yaml:"tables"`
	Roles       map[string]Role        `yaml:"roles"`
	Queries     map[string]Query       `yaml:"queries"`
	Indexes     map[string]Index[I]    `yaml:"indexes"`
}

type Table[T GType, C GConstraint] struct {
	Columns map[string]Column[T, C] `yaml:"columns"`
}

type Column[T GType, C GConstraint] struct {
	Type        T   `yaml:"type"`
	Constraints []C `yaml:"constraints"`
}

type Index[I GIndex] struct {
	Table  string `yaml:"table"`
	Column string `yaml:"column"`
	Using  I      `yaml:"using"`
}

type Role struct {
	Queries []string `yaml:"queries"`
}

type Query struct {
}
