package dba

type SQLStructure struct {
	ParameterizedQueries []ParameterizedQuery     `json:"parameterized_queries" yaml:"parameterized_queries" toml:"parameterized_queries" mapstructure:"parameterized_queries"`
	Roles                []Role                   `json:"roles" yaml:"roles" toml:"roles" mapstructure:"roles"`
	Tables               []Table                  `json:"tables" yaml:"tables" toml:"tables" mapstructure:"tables"`
	MappedConstraints    []map[string]interface{} `json:"constraints" yaml:"constraints" toml:"constraints" mapstructure:"constraints"`
	Constraints          []Constraint             `json:"-" yaml:"-" toml:"-" mapstructure:"-"`
}

type ParameterizedQuery struct {
	Name       string `json:"name" yaml:"name" toml:"name" mapstructure:"name"`
	Query      string `json:"query" yaml:"query" toml:"query" mapstructure:"query"`
	ReadOnly   bool   `json:"read_only" yaml:"read_only" toml:"read_only" mapstructure:"read_only"`
	Parameters interface{}
}

type Paramater struct {
	Name string `json:"name" yaml:"name" toml:"name" mapstructure:"name"`
	Type string `json:"type" yaml:"type" toml:"type" mapstructure:"type"`
}

type Table struct {
	Name       string      `json:"name" yaml:"name" toml:"name" mapstructure:"name"`
	Schema     string      `json:"schema" yaml:"schema" toml:"schema" mapstructure:"schema"`
	PrimaryKey string      `json:"primary_key" yaml:"primary_key" toml:"primary_key" mapstructure:"primary_key"`
	Columns    []ColumnDef `json:"columns" yaml:"columns" toml:"columns" mapstructure:"columns"`
}

type ColumnDef struct {
	Name       string                 `json:"name" yaml:"name" toml:"name" mapstructure:"name"`
	Nullable   bool                   `json:"nullable" yaml:"nullable" toml:"nullable" mapstructure:"nullable"`
	ColumnType map[string]interface{} `json:"column_type" yaml:"column_type" toml:"column_type" mapstructure:"column_type"`
}

type ColumnRef struct {
	Schema string `json:"schema" yaml:"schema" toml:"schema" mapstructure:"schema"`
	Table  string `json:"table" yaml:"table" toml:"table" mapstructure:"table"`
	Column string `json:"column" yaml:"column" toml:"column" mapstructure:"column"`
}

type ColumnType interface {
	GetType() string
}

type SqlDatabaseConfig struct {
	Name        string       `json:"name" yaml:"name" toml:"name" mapstructure:"name"`
	Owner       string       `json:"owner" yaml:"owner" toml:"owner" mapstructure:"owner"`
	DBType      string       `json:"type" yaml:"type" toml:"type" mapstructure:"type"`
	DefaultRole string       `json:"default_role" yaml:"default_role" toml:"default_role" mapstructure:"default_role"`
	Structure   SQLStructure `json:"structure" yaml:"structure" toml:"structure" mapstructure:"structure"`
}

func (s *SqlDatabaseConfig) GetName() string {
	return s.Name
}

func (s *SqlDatabaseConfig) GetOwner() string {
	return s.Owner
}

func (s *SqlDatabaseConfig) GetDBType() string {
	return s.DBType
}

func (s *SqlDatabaseConfig) GetDefaultRole() string {
	return s.DefaultRole
}

func (s *SqlDatabaseConfig) GetStructure() Structure {
	return &s.Structure
}
