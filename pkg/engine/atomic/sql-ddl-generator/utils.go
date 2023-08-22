package sqlddlgenerator

import (
	"fmt"
	"reflect"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

func columnTypeToSQLiteType(columnType types.DataType) (string, error) {
	err := columnType.Clean()
	if err != nil {
		return "", err
	}

	var sqlType string
	switch columnType {
	case types.TEXT:
		sqlType = "TEXT"
	case types.INT:
		sqlType = "INTEGER"
	case types.NULL:
		sqlType = "NULL"
	default:
		return "", fmt.Errorf("unknown column type: %s", columnType)
	}

	return sqlType, nil
}

func attributeToSQLiteString(colName string, attr *types.Attribute) (string, error) {
	formattedVal, err := formatAttributeValue(attr.Value)
	if err != nil {
		return "", err
	}

	err = attr.Clean()
	if err != nil {
		return "", err
	}

	switch attr.Type {
	case types.PRIMARY_KEY:
		return "", nil
	case types.DEFAULT:
		return "DEFAULT " + formattedVal, nil
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

func formatAttributeValue(value any) (string, error) {
	if value == nil {
		return "", nil
	}

	ref := reflect.ValueOf(value)

	switch ref.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", value), nil
	case reflect.String:
		return fmt.Sprintf("'%s'", value), nil
	default:
		return "", fmt.Errorf("unsupported value type %s", value)
	}
}
