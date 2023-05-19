package sqlddlgenerator

import (
	"fmt"
	"reflect"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
)

func columnTypeToSQLiteType(columnType dto.DataType) string {
	switch columnType {
	case dto.TEXT:
		return "TEXT"
	case dto.INT:
		return "INTEGER"
	default:
		return ""
	}
}

func attributeToSQLiteString(colName string, attr *dto.Attribute) (string, error) {
	formattedVal, err := formatAttributeValue(attr.Value)
	if err != nil {
		return "", err
	}

	switch attr.Type {
	case dto.PRIMARY_KEY:
		return "PRIMARY KEY", nil
	case dto.DEFAULT:
		return "DEFAULT " + formattedVal, nil
	case dto.NOT_NULL:
		return "NOT NULL", nil
	case dto.UNIQUE:
		return "UNIQUE", nil
	case dto.MIN:
		return "CHECK (" + colName + " >= " + formattedVal + ")", nil
	case dto.MAX:
		return "CHECK (" + colName + " <= " + formattedVal + ")", nil
	case dto.MIN_LENGTH:
		return "CHECK (LENGTH(" + colName + ") >= " + formattedVal + ")", nil
	case dto.MAX_LENGTH:
		return "CHECK (LENGTH(" + colName + ") <= " + formattedVal + ")", nil
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
