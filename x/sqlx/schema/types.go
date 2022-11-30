package schema

type KuniformColumn struct {
	Type       KuniformType                 `yaml:"type"`
	Attributes map[KuniformAttribute]string `yaml:"attributes"`
}

/*
type Attributes struct {
	MinLength int `yaml:"min_length"`
	MaxLength int `yaml:"max_length"`
	//Min        int    `yaml:"min"`
	//Max        int    `yaml:"max"`
	PrimaryKey bool   `yaml:"primary_key"`
	NotNull    bool   `yaml:"not_null" default:"false"`
	Unique     bool   `yaml:"unique"`
	Default    string `yaml:"default"`
}*/

type KuniformAttribute string

func (k KuniformAttribute) String() string {
	return string(k)
}

func (k KuniformAttribute) Valid() bool {
	return Attributes[k]
}

const (
	KuniformMinLength  KuniformAttribute = "min_length"
	KuniformMaxLength  KuniformAttribute = "max_length"
	KuniformPrimaryKey KuniformAttribute = "primary_key"
	KuniformNotNull    KuniformAttribute = "not_null"
	KuniformUnique     KuniformAttribute = "unique"
	KuniformDefault    KuniformAttribute = "default"
)

var Attributes = map[KuniformAttribute]bool{
	KuniformMinLength:  true,
	KuniformMaxLength:  true,
	KuniformPrimaryKey: true,
	KuniformNotNull:    true,
	KuniformUnique:     true,
	KuniformDefault:    true,
}

type KuniformAttributeType int

const (
	KATNull KuniformAttributeType = iota
	KATString
	KATInt
	KATBool
)

var AttributeTypes = map[KuniformAttribute]KuniformAttributeType{
	KuniformMinLength:  KATInt,
	KuniformMaxLength:  KATInt,
	KuniformPrimaryKey: KATBool,
	KuniformNotNull:    KATBool,
	KuniformUnique:     KATBool,
	KuniformDefault:    KATString,
}

type KuniformType string

func (k KuniformType) String() string {
	return string(k)
}

func (k KuniformType) Valid() bool {
	pt := Types[k]
	return pt.String() != ""
}

const (
	KuniformNull     KuniformType = "null"
	KuniformInt32    KuniformType = "int32"
	KuniformInt64    KuniformType = "int64"
	KuniformString   KuniformType = "string"
	KuniformBool     KuniformType = "bool"
	KuniformDatetime KuniformType = "datetime"
)

type PGType string

func (p PGType) String() string {
	return string(p)
}

const (
	PGNull     PGType = "null"
	PGInt32    PGType = "int4"
	PGInt64    PGType = "int8"
	PGString   PGType = "text"
	PGBool     PGType = "boolean"
	PGDatetime PGType = "timestamp"
)

type KuniformIndex string

func (k KuniformIndex) String() string {
	return string(k)
}

func (k KuniformIndex) Valid() bool {
	ind := Indexes[k]
	return ind.String() != ""
}

type PGIndex string

func (p PGIndex) String() string {
	return string(p)
}

const (
	KuniformBtree KuniformIndex = "btree"
)

const (
	PGBtree PGIndex = "btree"
)

var Types = map[KuniformType]PGType{
	KuniformNull:     PGNull,
	KuniformInt32:    PGInt32,
	KuniformInt64:    PGInt64,
	KuniformString:   PGString,
	KuniformBool:     PGBool,
	KuniformDatetime: PGDatetime,
}

var Indexes = map[KuniformIndex]PGIndex{
	KuniformBtree: PGBtree,
}
