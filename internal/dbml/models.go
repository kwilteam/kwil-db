package dbml

type RelationshipType string

const (
	None       RelationshipType = ""
	OneToOne   RelationshipType = "OneToOne"
	OneToMany  RelationshipType = "OneToMany"
	ManyToOne  RelationshipType = "ManyToOne"
	ManyToMany RelationshipType = "ManyToMany"
)

type DBML struct {
	Project     Project      `json:"project,omitempty"`
	Tables      []Table      `json:"tables,omitempty"`
	Enums       []Enum       `json:"enums,omitempty"`
	Refs        []Ref        `json:"refs,omitempty"`
	Queries     []Query      `json:"queries,omitempty"`
	Roles       []Role       `json:"roles,omitempty"`
	TableGroups []TableGroup `json:"tableGroups,omitempty"`
}

type Project struct {
	Name         string `json:"name,omitempty"`
	Note         string `json:"note,omitempty"`
	DatabaseType string `json:"databaseType,omitempty"`
}

type Table struct {
	Name    string   `json:"name,omitempty"`
	Alias   string   `json:"alias,omitempty"`
	Columns []Column `json:"columns,omitempty"`
	Indexes []Index  `json:"indexes,omitempty"`
	Note    string   `json:"note,omitempty"`
}

type Query struct {
	Name       string `json:"name,omitempty"`
	Expression string `json:"expression,omitempty"`
}

type Role struct {
}

type Column struct {
	Name     string        `json:"name,omitempty"`
	Type     string        `json:"type,omitempty"`
	Size     int           `json:"size,omitempty"`
	Settings ColumnSetting `json:"settings"`
}

type ColumnSetting struct {
	Note          string     `json:"note,omitempty"`
	PK            bool       `json:"pk,omitempty"`
	Unique        bool       `json:"unique,omitempty"`
	Default       string     `json:"default,omitempty"`
	NotNull       bool       `json:"nullable,omitempty"`
	AutoIncrement bool       `json:"autoIncrement,omitempty"`
	Array         bool       `json:"array,omitempty"`
	Unsigned      bool       `json:"unsigned,omitempty"`
	Ref           *OneWayRef `json:"ref,omitempty"`
}

type OneWayRef struct {
	To   Rel              `json:"to,omitempty"`
	Type RelationshipType `json:"type,omitempty"`
}

type Rel struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

type Index struct {
	Fields   []string     `json:"fields,omitempty"`
	Settings IndexSetting `json:"settings,omitempty"`
}

type IndexSetting struct {
	Type   string `json:"type,omitempty"`
	Name   string `json:"name,omitempty"`
	Unique bool   `json:"unique,omitempty"`
	PK     bool   `json:"pk,omitempty"`
	Note   string `json:"note,omitempty"`
}

type Relationship struct {
	From Rel              `json:"from,omitempty"`
	To   Rel              `json:"to,omitempty"`
	Type RelationshipType `json:"type,omitempty"`
}

var RelationshipMap = map[Token]RelationshipType{
	GT:   ManyToOne,
	LT:   OneToMany,
	SUB:  OneToOne,
	LTGT: ManyToMany,
}

type Ref struct {
	Name          string         `json:"name,omitempty"`
	Relationships []Relationship `json:"relationships,omitempty"`
}

type Enum struct {
	Name   string      `json:"name,omitempty"`
	Values []EnumValue `json:"values,omitempty"`
}

type EnumValue struct {
	Name string `json:"name,omitempty"`
	Note string `json:"note,omitempty"`
}

type TableGroup struct {
	Name    string   `json:"name,omitempty"`
	Members []string `json:"members,omitempty"`
}
