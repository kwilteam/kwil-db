package ast

import (
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/token"
)

var columnTypes = map[string]types.DataType{
	"text": types.TEXT,
	"int":  types.INT,
}

var attributeTypes = map[token.Token]types.AttributeType{
	token.UNIQUE:  types.UNIQUE,
	token.PRIMARY: types.PRIMARY_KEY,
	token.NOTNULL: types.NOT_NULL,
	token.MINLEN:  types.MIN_LENGTH,
	token.MAXLEN:  types.MAX_LENGTH,
	token.MIN:     types.MIN,
	token.MAX:     types.MAX,
	token.DEFAULT: types.DEFAULT,
}

func GetMappedColumnType(typeName string) types.DataType {
	dataType, ok := columnTypes[typeName]
	if !ok {
		dataType = types.INVALID_DATA_TYPE
	}
	return dataType
}

func GetMappedAttributeType(t token.Token) types.AttributeType {
	attributeType, ok := attributeTypes[t]
	if !ok {
		attributeType = types.INVALID_ATTRIBUTE_TYPE
	}
	return attributeType
}
