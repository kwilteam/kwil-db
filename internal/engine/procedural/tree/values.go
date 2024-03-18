package tree

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	coreTypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/procedural/types"
)

// a Value is a variable, text, int, boolean, uuid, text[], composite type, etc.
type Value interface {
	PGMarshaler
	Expression
	value() // private method to ensure that only values can be assigned to a Value

}

// Variable is a reference to a value.
// In the language, they are prefixed with a $.
// Once parsed, the $ is removed.
// It can be variables such as $var, or $row.field1.field2
type Variable struct {
	// Name is the name of the variable.
	// It is case-sensitive.
	// It does not include the $.
	// It should include all fields, separated by dots.
	Name string
}

func (Variable) value()      {}
func (Variable) expression() {}
func (v *Variable) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	d, ok := info.Context.Variables[v.Name]
	if !ok {
		return 0, nil, fmt.Errorf("variable %s not found", v.Name)
	}

	return ReturnTypeValue, d, nil
}

// MarshalPG will return the name of the variable.
func (v *Variable) MarshalPG(info *SystemInfo) (string, error) {
	return v.Name, nil // TODO: we need to hash the first period to ensure uniqueness, since postgres
	// will complain if it matches a column name.
}

// TextValue is a string.
// It should not be wrapped in single quotes.
type TextValue string

func (TextValue) value()      {}
func (TextValue) expression() {}
func (v *TextValue) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	return ReturnTypeValue, types.TypeText, nil
}

func (t *TextValue) MarshalPG(info *SystemInfo) (string, error) {
	return fmt.Sprintf("'%s'", *t), nil
}

type IntValue int64

func (IntValue) value()      {}
func (IntValue) expression() {}
func (v *IntValue) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	return ReturnTypeValue, types.TypeInt, nil
}

func (i *IntValue) MarshalPG(info *SystemInfo) (string, error) {
	return strconv.FormatInt(int64(*i), 10), nil
}

// type Uint256Value big.Int

// func (Uint256Value) value()      {}
// func (Uint256Value) expression() {}
// func (v *Uint256Value) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
// 	return ReturnTypeValue, types.Type, nil
// }

// func (u *Uint256Value) MarshalPG(info *SystemInfo) (string, error) {
// 	b := big.Int(*u)
// 	return b.String(), nil
// }

type BooleanValue bool

func (BooleanValue) value()      {}
func (BooleanValue) expression() {}
func (v *BooleanValue) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	return ReturnTypeValue, types.TypeBoolean, nil
}

func (b *BooleanValue) MarshalPG(info *SystemInfo) (string, error) {
	return strconv.FormatBool(bool(*b)), nil
}

type UUIDValue coreTypes.UUID

func (UUIDValue) value()      {}
func (UUIDValue) expression() {}
func (v *UUIDValue) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	return ReturnTypeValue, types.TypeUUID, nil
}

func (u *UUIDValue) MarshalPG(info *SystemInfo) (string, error) {
	id := coreTypes.UUID(*u)
	return fmt.Sprintf("'%s'", id.String()), nil
}

type BlobValue []byte

func (BlobValue) value()      {}
func (BlobValue) expression() {}
func (v *BlobValue) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	return ReturnTypeValue, types.TypeBlob, nil
}

// MarshalPG will convert the blob to a hex string, escape it, and wrap it in single quotes.
func (b *BlobValue) MarshalPG(info *SystemInfo) (string, error) {
	hexStr := hex.EncodeToString(*b)
	return fmt.Sprintf(`'\x%s'`, hexStr), nil
}

type ArrayValue struct {
	DataType types.DataType // the type of the array
	Values   []Value
}

func (ArrayValue) value()      {}
func (ArrayValue) expression() {}
func (v *ArrayValue) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	return ReturnTypeValue, v.DataType, nil
}

func (a *ArrayValue) MarshalPG(info *SystemInfo) (string, error) {
	str := strings.Builder{}
	str.WriteString("ARRAY[")
	for i, v := range a.Values {
		if i > 0 {
			str.WriteString(", ")
		}
		marshaled, err := v.MarshalPG(info)
		if err != nil {
			return "", err
		}
		str.WriteString(marshaled)
	}
	str.WriteString("]")
	return str.String(), nil
}

type CompositeValue struct {
	// Type is the definition of the composite type.
	Type *types.CustomType
	// Values are the fields of the composite value.
	// They are passed as:
	// {
	//   "field1": value1,
	//   "field2": value2,
	// }
	Values map[string]Value
}

func (CompositeValue) value()      {}
func (CompositeValue) expression() {}
func (v *CompositeValue) assigns(info *SystemInfo) (ReturnType, types.DataType, error) {
	return ReturnTypeValue, v.Type, nil
}

// MarshalPG marshals the composite value into a string that can be used in a SQL statement.
func (c *CompositeValue) MarshalPG(info *SystemInfo) (string, error) {
	// To marshall to postgres, we will use postgres's row constructor syntax
	// with type casting.
	str := strings.Builder{}
	str.WriteString("ROW(")

	// get the definition for the composite type
	schemaName := c.Type.Schema
	if schemaName == "" {
		schemaName = "public"
	}
	schema, ok := info.Schemas[schemaName]
	if !ok {
		return "", fmt.Errorf("schema %s not found", schemaName)
	}

	customType, ok := schema.Types[c.Type.Name]
	if !ok {
		return "", fmt.Errorf("type %s not found in schema %s", c.Type.Name, schemaName)
	}

	// match up the keys in the values with the order of the fields in the type
	for i, field := range customType.Fields {
		val, ok := c.Values[field.Name]
		if !ok {
			return "", fmt.Errorf("field %s is missing", field.Name)
		}

		if i > 0 && i < len(customType.Fields) {
			str.WriteString(", ")
		}

		marshaled, err := val.MarshalPG(info)
		if err != nil {
			return "", err
		}

		str.WriteString(marshaled)
	}

	str.WriteString(")::")
	str.WriteString(schemaName)
	str.WriteString(".")
	str.WriteString(c.Type.Name)
	return str.String(), nil
}
