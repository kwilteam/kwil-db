package sqlddlgenerator

import (
	"fmt"
	"reflect"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
)

func columnTypeToSQLiteType(columnType dto.DataType) (string, error) {
	err := columnType.Clean()
	if err != nil {
		return "", err
	}

	sqlType := ""
	switch columnType {
	case dto.TEXT:
		sqlType = "TEXT"
	case dto.INT:
		sqlType = "INTEGER"
	case dto.NULL:
		sqlType = "NULL"
	default:
		return "", fmt.Errorf("unknown column type: %s", columnType)
	}

	return sqlType, nil
}

func attributeToSQLiteString(colName string, attr *dto.Attribute) (string, error) {
	formattedVal, err := formatAttributeValue(attr.Value)
	if err != nil {
		return "", err
	}

	err = attr.Clean()
	if err != nil {
		return "", err
	}

	switch attr.Type {
	case dto.PRIMARY_KEY:
		return "", nil
	case dto.DEFAULT:
		return "DEFAULT " + formattedVal, nil
	case dto.NOT_NULL:
		return "NOT NULL", nil
	case dto.UNIQUE:
		return "UNIQUE", nil
	case dto.MIN:
		intVal, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		return "CHECK (" + colName + " >= " + fmt.Sprint(intVal) + ")", nil
	case dto.MAX:
		intVal, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		return "CHECK (" + colName + " <= " + fmt.Sprint(intVal) + ")", nil
	case dto.MIN_LENGTH:
		intVal, err := conv.Int(attr.Value)
		if err != nil {
			return "", err
		}

		return "CHECK (LENGTH(" + colName + ") >= " + fmt.Sprint(intVal) + ")", nil
	case dto.MAX_LENGTH:
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
