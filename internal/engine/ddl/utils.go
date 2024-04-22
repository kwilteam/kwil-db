package ddl

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/conv"
)

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
