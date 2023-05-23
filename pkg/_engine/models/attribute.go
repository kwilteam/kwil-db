package models

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

type Attribute struct {
	Type  types.AttributeType `json:"type" clean:"is_enum,attribute_type"`
	Value []byte              `json:"value"`
}

// Coerce will coerce the attribute value to the correct data type, depending on the attribute type.
// It takes an input of a column type, which is used in the case that the attribute type is DEFAULT
func (a *Attribute) Coerce(columnType types.DataType) error {
	switch a.Type {
	case types.PRIMARY_KEY, types.UNIQUE, types.NOT_NULL:
		a.Value = types.NewEmpty().Bytes()
	case types.DEFAULT:
		// default must be the same type as the column
		return a.assertType(columnType)
	case types.MIN, types.MAX, types.MIN_LENGTH, types.MAX_LENGTH:
		// min, max, min_length, max_length must be int, regardless of column type
		return a.assertType(types.INT)
	default:
		return fmt.Errorf("invalid attribute type: %d", a.Type)
	}
	return nil
}

// assertType will convert the attribute value to the correct serialized type if it is not already
func (a *Attribute) assertType(typ types.DataType) error {
	concVal, err := types.NewFromSerial(a.Value)
	if err != nil {
		return err
	}

	if concVal.Type() == typ {
		return nil
	}

	newConcVal, err := typ.Coerce(concVal)
	if err != nil {
		return err
	}

	a.Value = newConcVal.Bytes()
	return nil
}
