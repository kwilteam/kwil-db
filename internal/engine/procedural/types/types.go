// package types contains the types used in the procedure execution.
// this will probably be moved to a different package in the future.
// These will be parsed out by the Kuneiform parser.
package types

import "fmt"

// CompositeTypeDefinition is a user-defined type.
type CompositeTypeDefinition struct {
	// Name is the name of the type.
	// It should always be lower case.
	Name string
	// Fields are the fields of the composite type.
	Fields []*CompositeTypeField
}

// CompositeTypeField is a field of a composite type.
type CompositeTypeField struct {
	// Name is the name of the parameter.
	// It should always be lower case.
	Name string
	// Type is the type of the parameter.
	Type DataType
}

// DataType is any data type.
type DataType interface {
	String() string
	Equals(DataType) bool
	returns() // ensure that only types in this package can be return types
}

// CustomType is a type defined in a schema that is not a built-in type.
type CustomType struct {
	// Schema is the schema of the type, if it is a composite type.
	// If it is a built-in type, this is empty.
	Schema string
	// Name is the name of the type.
	// It should always be lower case.
	Name string
}

// String returns the string representation of the type.
func (t *CustomType) String() string {
	if t.Schema == "" {
		return t.Name
	}
	return t.Schema + "." + t.Name
}

func (c *CustomType) Equals(d DataType) bool {
	if d, ok := d.(*CustomType); ok {
		return c.Schema == d.Schema && c.Name == d.Name
	}
	return false
}

func (CustomType) returns() {}

// ArrayType is an array of some type.
type ArrayType struct {
	// Type is the type of the array.
	Type DataType
}

// String returns the string representation of the type.
func (t *ArrayType) String() string {
	return fmt.Sprintf("%s[]", t.Type.String())
}

func (ArrayType) Equals(d DataType) bool {
	arr, ok := d.(*ArrayType)
	if !ok {
		return false
	}
	return arr.Type.Equals(arr.Type)
}
func (ArrayType) returns() {}

// DefaultType is the default type.
type DefaultType string

// String returns the string representation of the type.
func (t DefaultType) String() string {
	return string(t)
}

func (c DefaultType) Equals(d DataType) bool {
	if d, ok := d.(DefaultType); ok {
		return c == d
	}
	return false
}
func (DefaultType) returns() {}

const (
	TypeText    DefaultType = "text"
	TypeInt     DefaultType = "int8"
	TypeBoolean DefaultType = "boolean"
	TypeBlob    DefaultType = "bytea"
	TypeUUID    DefaultType = "uuid"
)
