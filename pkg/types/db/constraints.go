package db

type Constraint interface {
	GetName() string
	GetType() string
}

type ForeignKeyConstraint struct {
	Name      string    `json:"name" yaml:"name" toml:"name" mapstructure:"name"`
	Type      string    `json:"type" yaml:"type" toml:"type" mapstructure:"type"`
	Key       ColumnRef `json:"key" yaml:"key" toml:"key" mapstructure:"key"`
	Reference ColumnRef `json:"reference" yaml:"reference" toml:"reference" mapstructure:"reference"`
}

func (f ForeignKeyConstraint) GetName() string {
	return f.Name
}

func (f ForeignKeyConstraint) GetType() string {
	return f.Type
}
