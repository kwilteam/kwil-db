package sqlitegenerator

import (
	"fmt"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/types"
)

func columnTypeToSQLiteType(columnType types.DataType) string {
	switch columnType {
	case types.TEXT:
		return "TEXT"
	case types.INT:
		return "INTEGER"
	default:
		return ""
	}
}

func attributeToSQLiteString(colName string, attr *models.Attribute) (string, error) {
	val := types.NewEmpty()
	formattedVal := ""
	var err error
	if attr.Value != nil {
		val, err = types.NewFromSerial(attr.Value)
		if err != nil {
			return "", err
		}

		formattedVal, err = formatAttributeValue(val)
		if err != nil {
			return "", err
		}
	}

	switch attr.Type {
	case types.PRIMARY_KEY:
		return "PRIMARY KEY", nil
	case types.DEFAULT:
		if val.IsEmpty() {
			return "", fmt.Errorf("default value cannot be empty")
		}
		return "DEFAULT " + formattedVal, nil
	case types.NOT_NULL:
		return "NOT NULL", nil
	case types.UNIQUE:
		return "UNIQUE", nil
	case types.MIN:
		return "CHECK (" + colName + " >= " + formattedVal + ")", nil
	case types.MAX:
		return "CHECK (" + colName + " <= " + formattedVal + ")", nil
	case types.MIN_LENGTH:
		return "CHECK (LENGTH(" + colName + ") >= " + formattedVal + ")", nil
	case types.MAX_LENGTH:
		return "CHECK (LENGTH(" + colName + ") <= " + formattedVal + ")", nil
	default:
		return "", nil
	}
}

func formatAttributeValue(value *types.ConcreteValue) (string, error) {
	switch value.Type() {
	case types.INT:
		str, err := value.AsString()
		if err != nil {
			return "", err
		}
		return str, nil
	case types.TEXT:
		str, err := value.AsString()
		if err != nil {
			return "", err
		}
		return "'" + str + "'", nil
	case types.NULL:
		return "", nil
	default:
		return "", fmt.Errorf("unknown attribute value type %s", value.Type().String())
	}
}
