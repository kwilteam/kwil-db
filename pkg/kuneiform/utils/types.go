package utils

import (
	"strings"

	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/token"
)

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

func GetMappedIndexType(typeName string) schema.IndexType {
	return schema.IndexType(typeName)
}

func GetMappedColumnType(typeName string) schema.ColumnType {
	return schema.ColumnType(strings.ToUpper(typeName))
}

func GetMappedAttributeType(t token.Token) schema.AttributeType {
	return schema.AttributeType(attributeTypes[t])
}
