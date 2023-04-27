package schema

// ColumnType is the type of column
type ColumnType int

const (
	InvalidColumnType ColumnType = iota + 100
	ColNull
	ColText
	ColInt
	EndColumnType
)

var columnTypes = [...]string{
	ColNull: "null",
	ColText: "text",
	ColInt:  "int",
}

func (t ColumnType) String() string {
	if t <= InvalidColumnType || t >= EndColumnType {
		return "unknown"
	}
	return columnTypes[t]
}

// AttributeType is the type of attribute
type AttributeType int

const (
	InvalidAttributeType AttributeType = iota + 100
	AttrPrimaryKey
	AttrUnique
	AttrNotNull
	AttrDefault
	AttrMin       // Min allowed value
	AttrMax       // Max allowed value
	AttrMinLength // Min allowed length
	AttrMaxLength // Max allowed length
	EndAttributeType
)

var attributeTypes = [...]string{
	AttrPrimaryKey: "primary_key",
	AttrUnique:     "unique",
	AttrNotNull:    "not_null",
	AttrDefault:    "default",
	AttrMin:        "min",
	AttrMax:        "max",
	AttrMinLength:  "min_length",
	AttrMaxLength:  "max_length",
}

func (t AttributeType) String() string {
	if t <= InvalidAttributeType || t >= EndAttributeType {
		return "unknown"
	}
	return attributeTypes[t]
}

// IndexType is the type of index
type IndexType int

const (
	InvalidIndexType IndexType = iota + 100
	IdxBtree
	IdxUniqueBtree
	EndIndexType
)

var indexTypes = [...]string{
	IdxBtree:       "btree",
	IdxUniqueBtree: "unique",
}

func (t IndexType) String() string {
	if t <= InvalidIndexType || t >= EndIndexType {
		return "unknown"
	}
	return indexTypes[t]
}
