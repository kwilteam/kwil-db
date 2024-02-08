package ddl

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/conv"
	"github.com/kwilteam/kwil-db/internal/engine/types"
)

func columnTypeToSQLType(columnType types.DataType) (string, error) {
	err := columnType.Clean()
	if err != nil {
		return "", err
	}

	var sqlType string
	switch columnType {
	case types.TEXT:
		sqlType = "TEXT"
	case types.INT:
		sqlType = "INT8"
	case types.NULL:
		sqlType = "NULL"
	case types.BLOB:
		sqlType = "BYTEA"
	case types.BOOL:
		sqlType = "BOOL"
	default:
		return "", fmt.Errorf("unknown column type: %s", columnType)
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
