package db

type Role struct {
	Name        string `json:"name" yaml:"name" toml:"name" mapstructure:"name"`
	Permissions Permissions
}

type Permissions struct {
	DDL                  bool     `json:"ddl" yaml:"ddl" toml:"ddl" mapstructure:"ddl"`
	ParamaterizedQueries []string `json:"parameterized_queries" yaml:"parameterized_queries" toml:"parameterized_queries" mapstructure:"parameterized_queries"`
}

func (r *Role) GetName() string {
	return r.Name
}

func (r *Role) GetPermissions() Permissions {
	return r.Permissions
}
