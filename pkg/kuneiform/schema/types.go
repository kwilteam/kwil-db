package schema

import (
	"fmt"
	"strings"
)

// ColumnType is the type of column
type ColumnType string

const (
	ColNull ColumnType = "null"
	ColText ColumnType = "text"
	ColInt  ColumnType = "int"
)

func (t ColumnType) String() string {
	return string(t)
}

func GetColumnType(col string) (ColumnType, error) {
	column, ok := columnTypes[strings.ToLower(col)]
	if !ok {
		return "", fmt.Errorf("invalid column type: %s", col)
	}

	return column, nil
}

var columnTypes = map[string]ColumnType{
	"null": ColNull,
	"text": ColText,
	"int":  ColInt,
}

// AttributeType is the type of attribute
type AttributeType string

const (
	AttrPrimaryKey AttributeType = "primary_key"
	AttrUnique     AttributeType = "unique"
	AttrNotNull    AttributeType = "not_null"
	AttrDefault    AttributeType = "default"
	AttrMin        AttributeType = "min"
	AttrMax        AttributeType = "max"
	AttrMinLength  AttributeType = "min_length"
	AttrMaxLength  AttributeType = "max_length"
)

func (t AttributeType) String() string {
	return string(t)
}

func GetAttributeType(attr string) (AttributeType, error) {
	attribute, ok := attributeTypes[strings.ToLower(attr)]
	if !ok {
		return "", fmt.Errorf("invalid attribute type: %s", attr)
	}

	return attribute, nil
}

// attributeTypes maps the Kuneiform attribute tokens to the schema attribute type
var attributeTypes = map[string]AttributeType{
	"primary": AttrPrimaryKey,
	"unique":  AttrUnique,
	"notnull": AttrNotNull,
	"default": AttrDefault,
	"min":     AttrMin,
	"max":     AttrMax,
	"minlen":  AttrMinLength,
	"maxlen":  AttrMaxLength,
}

// IndexType is the type of index
type IndexType string

const (
	IdxBtree       IndexType = "btree"
	IdxUniqueBtree IndexType = "unique_btree"
	IdxPrimary     IndexType = "primary"
)

func (t IndexType) String() string {
	return string(t)
}

// indexTypes maps the Kuneiform index tokens to the schema index type
var indexTypes = map[string]IndexType{
	"index":   IdxBtree,
	"unique":  IdxUniqueBtree,
	"primary": IdxPrimary,
}

func GetIndexType(idxType string) (IndexType, error) {
	index, ok := indexTypes[strings.ToLower(idxType)]
	if !ok {
		return "", fmt.Errorf("invalid index type: %s", idxType)
	}

	return index, nil
}
