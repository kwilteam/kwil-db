package ddl

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/conv"
)

func columnTypeToSQLType(columnType *types.DataType) (string, error) {
	err := columnType.Clean()
	if err != nil {
		return "", err
	}

	var sqlType string
	switch columnType {
	case types.TextType:
		sqlType = "TEXT"
	case types.IntType:
		sqlType = "INT8"
	case types.NullType:
		return "", fmt.Errorf("cannot have null column type")
	case types.BlobType:
		sqlType = "BYTEA"
	case types.BoolType:
		sqlType = "BOOLEAN"
	case types.UUIDType:
		sqlType = "UUID"
	default:
		// based on an alias
		sqlType = columnType.String()
	}

	return sqlType, nil
}

func attributeToSQLString(colName string, attr *types.Attribute) (string, error) {
	err := attr.Clean()
	if err != nil {
		return "", err
	}

	switch attr.Type {
	case types.PRIMARY_KEY:
		return "", nil
	case types.DEFAULT:
		return "DEFAULT " + attr.Value, nil
	case types.NOT_NULL:
		return "NOT NULL", nil
	case types.UNIQUE:
		return "UNIQUE", nil
	case types.MIN:
		intVal, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		return "CHECK (" + colName + " >= " + fmt.Sprint(intVal) + ")", nil
	case types.MAX:
		intVal, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		return "CHECK (" + colName + " <= " + fmt.Sprint(intVal) + ")", nil
	case types.MIN_LENGTH:
		intVal, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		return "CHECK (LENGTH(" + colName + ") >= " + fmt.Sprint(intVal) + ")", nil
	case types.MAX_LENGTH:
		intVal, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		return "CHECK (LENGTH(" + colName + ") <= " + fmt.Sprint(intVal) + ")", nil
	default:
		return "", nil
	}
}
