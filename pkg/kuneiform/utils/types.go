package utils

import (
	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/token"
	"strings"
)

var columnTypes = map[string]schema.ColumnType{
	"text": schema.ColText,
	"int":  schema.ColInt,
}

var attributeTypes = map[token.Token]schema.AttributeType{
	token.UNIQUE:  schema.AttrUnique,
	token.PRIMARY: schema.AttrPrimaryKey,
	token.NOTNULL: schema.AttrNotNull,
	token.MINLEN:  schema.AttrMinLength,
	token.MAXLEN:  schema.AttrMaxLength,
	token.MIN:     schema.AttrMin,
	token.MAX:     schema.AttrMax,
	token.DEFAULT: schema.AttrDefault,
}

var indexTypes = map[string]schema.IndexType{
	"btree":  schema.IdxBtree,
	"unique": schema.IdxUniqueBtree,
}

func GetMappedIndexType(typeName string) schema.IndexType {
	typeName = strings.ToLower(typeName)
	indexType, ok := indexTypes[typeName]
	if !ok {
		indexType = schema.InvalidIndexType
	}
	return indexType
}

func GetMappedColumnType(typeName string) schema.ColumnType {
	typeName = strings.ToLower(typeName)
	dataType, ok := columnTypes[typeName]
	if !ok {
		dataType = schema.InvalidColumnType
	}
	return dataType
}

func GetMappedAttributeType(t token.Token) schema.AttributeType {
	attributeType, ok := attributeTypes[t]
	if !ok {
		attributeType = schema.InvalidAttributeType
	}
	return attributeType
}
