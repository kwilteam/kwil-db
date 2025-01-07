package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/core/types/decimal"
)

// DataType is a data type.
// It includes both built-in types and user-defined types.
type DataType struct {
	// Name is the name of the type.
	Name string `json:"name"`
	// IsArray is true if the type is an array.
	IsArray bool `json:"is_array"`
	// Metadata is the metadata of the type.
	Metadata [2]uint16 `json:"metadata"`
}

func (c DataType) SerializeSize() int {
	// uint16 version + uint32 name length + name + uint8 is_array +
	//   2 x uint16 metadata
	return 2 + 4 + len(c.Name) + 1 + 4
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func (c DataType) MarshalBinary() ([]byte, error) {
	b := make([]byte, c.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2
	binary.BigEndian.PutUint32(b[offset:], uint32(len(c.Name)))
	offset += 4
	copy(b[offset:], c.Name)
	offset += len(c.Name)
	b[offset] = boolToByte(c.IsArray)
	offset++
	binary.BigEndian.PutUint16(b[offset:], c.Metadata[0])
	offset += 2
	binary.BigEndian.PutUint16(b[offset:], c.Metadata[1])
	return b, nil
}

func (c *DataType) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid tuple data, unknown version %d", ver)
	}
	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen+1+2*2 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	c.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	switch data[offset] {
	case 0:
	case 1:
		c.IsArray = true
	default:
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	offset++

	c.Metadata[0] = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	c.Metadata[1] = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	if offset != c.SerializeSize() { // bug, must match
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	return nil
}

// String returns the string representation of the type.
func (c *DataType) String() string {
	str := strings.Builder{}
	str.WriteString(c.Name)
	if c.IsArray {
		return str.String() + "[]"
	}

	if c.Name == DecimalStr {
		str.WriteString("(")
		str.WriteString(strconv.FormatUint(uint64(c.Metadata[0]), 10))
		str.WriteString(",")
		str.WriteString(strconv.FormatUint(uint64(c.Metadata[1]), 10))
		str.WriteString(")")
	}

	return str.String()
}

func (c *DataType) HasMetadata() bool {
	return c.Metadata != ZeroMetadata
}

var ZeroMetadata = [2]uint16{}

// PGString returns the string representation of the type in Postgres.
func (c *DataType) PGString() (string, error) {
	scalar, err := c.PGScalar()
	if err != nil {
		return "", err
	}

	if c.IsArray {
		return scalar + "[]", nil
	}

	return scalar, nil
}

// PGScalar returns the scalar representation of the type in Postgres.
// For example, if this is of type DECIMAL(100,5)[], it will return NUMERIC(100,5).
func (c *DataType) PGScalar() (string, error) {
	var scalar string
	switch strings.ToLower(c.Name) {
	case intStr:
		scalar = "INT8"
	case textStr:
		scalar = "TEXT"
	case boolStr:
		scalar = "BOOL"
	case blobStr:
		scalar = "BYTEA"
	case uuidStr:
		scalar = "UUID"
	case uint256Str:
		scalar = "UINT256"
	case DecimalStr:
		if c.Metadata == ZeroMetadata {
			scalar = "NUMERIC"
		} else {
			scalar = fmt.Sprintf("NUMERIC(%d,%d)", c.Metadata[0], c.Metadata[1])
		}
	case nullStr:
		return "", errors.New("cannot have null column type")
	case unknownStr:
		return "", errors.New("cannot have unknown column type")
	default:
		return "", fmt.Errorf("unknown column type: %s", c.Name)
	}

	return scalar, nil
}

func (c *DataType) Clean() error {
	lName := strings.ToLower(c.Name)

	referencedType, ok := typeAlias[lName]
	if !ok {
		return fmt.Errorf("unknown type: %s", c.Name)
	}

	switch referencedType {
	case intStr, textStr, boolStr, blobStr, uuidStr, uint256Str: // ok
		if c.Metadata != ZeroMetadata {
			return fmt.Errorf("type %s cannot have metadata", c.Name)
		}
	case DecimalStr:
		if c.Metadata != ZeroMetadata {
			err := decimal.CheckPrecisionAndScale(c.Metadata[0], c.Metadata[1])
			if err != nil {
				return err
			}
		}
	case nullStr, unknownStr:
		if c.IsArray {
			return fmt.Errorf("type %s cannot be an array", c.Name)
		}

		if c.Metadata != ZeroMetadata {
			return fmt.Errorf("type %s cannot have metadata", c.Name)
		}
	default:
		return fmt.Errorf("unknown type: %s", c.Name)
	}

	c.Name = referencedType

	return nil
}

// Copy returns a copy of the type.
func (c *DataType) Copy() *DataType {
	d := &DataType{
		Name:     c.Name,
		IsArray:  c.IsArray,
		Metadata: c.Metadata,
	}

	return d
}

// EqualsStrict returns true if the type is equal to the other type.
// The types must be exactly the same, including metadata.
func (c *DataType) EqualsStrict(other *DataType) bool {
	// if unknown, return true. unknown is a special case used
	// internally when type checking is disabled.
	if c.Name == unknownStr || other.Name == unknownStr {
		return true
	}

	if c.IsArray != other.IsArray {
		return false
	}

	if (c.Metadata == ZeroMetadata) != (other.Metadata == ZeroMetadata) {
		return false
	}
	if c.Metadata != ZeroMetadata {
		if c.Metadata[0] != other.Metadata[0] || c.Metadata[1] != other.Metadata[1] {
			return false
		}
	}

	return strings.EqualFold(c.Name, other.Name)
}

// Equals returns true if the type is equal to the other type, or if either type is null.
func (c *DataType) Equals(other *DataType) bool {
	if c.Name == nullStr || other.Name == nullStr {
		return true
	}

	return c.EqualsStrict(other)
}

func (c *DataType) IsNumeric() bool {
	if c.IsArray {
		return false
	}

	return c.Name == intStr || c.Name == DecimalStr || c.Name == uint256Str || c.Name == unknownStr
}

// declared DataType constants.
// We do not have one for fixed because fixed types require metadata.
var (
	IntType = &DataType{
		Name: intStr,
	}
	IntArrayType = ArrayType(IntType)
	TextType     = &DataType{
		Name: textStr,
	}
	TextArrayType = ArrayType(TextType)
	BoolType      = &DataType{
		Name: boolStr,
	}
	BoolArrayType = ArrayType(BoolType)
	BlobType      = &DataType{
		Name: blobStr,
	}
	BlobArrayType = ArrayType(BlobType)
	UUIDType      = &DataType{
		Name: uuidStr,
	}
	UUIDArrayType = ArrayType(UUIDType)
	// DecimalType contains 1,0 metadata.
	// For type detection, users should prefer compare a datatype
	// name with the DecimalStr constant.
	DecimalType = &DataType{
		Name:     DecimalStr,
		Metadata: [2]uint16{1, 0}, // the minimum precision and scale
	}
	DecimalArrayType = ArrayType(DecimalType)
	Uint256Type      = &DataType{
		Name: uint256Str,
	}
	Uint256ArrayType = ArrayType(Uint256Type)
	// NullType is a special type used internally
	NullType = &DataType{
		Name: nullStr,
	}
	// Unknown is a special type used internally
	// when a type is unknown until runtime.
	UnknownType = &DataType{
		Name: unknownStr,
	}
)

// ArrayType creates an array type of the given type.
// It panics if the type is already an array.
func ArrayType(t *DataType) *DataType {
	if t.IsArray {
		panic("cannot create an array of an array")
	}
	return &DataType{
		Name:     t.Name,
		IsArray:  true,
		Metadata: t.Metadata,
	}
}

const (
	textStr    = "text"
	intStr     = "int8"
	boolStr    = "bool"
	blobStr    = "blob"
	uuidStr    = "uuid"
	uint256Str = "uint256"
	// DecimalStr is a fixed point number.
	DecimalStr = "decimal"
	nullStr    = "null"
	unknownStr = "unknown"
)

// NewDecimalType creates a new fixed point decimal type.
func NewDecimalType(precision, scale uint16) (*DataType, error) {
	err := decimal.CheckPrecisionAndScale(precision, scale)
	if err != nil {
		return nil, err
	}

	return &DataType{
		Name:     DecimalStr,
		Metadata: [2]uint16{precision, scale},
	}, nil
}

// ParseDataType parses a string into a data type.
func ParseDataType(s string) (*DataType, error) {
	// four cases: TEXT, TEXT[], TEXT(1,2), TEXT(1,2)[]
	// we will parse the type first, then the array, then the metadata
	// we will not allow metadata without an array

	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return nil, errors.New("empty data type")
	}

	// Regular expression to parse the data type
	re := regexp.MustCompile(`^([a-z0-9]+)(\(([\d, ]+)\))?(\[\])?$`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid data type format: %s", s)
	}

	baseType := matches[1]
	rawMetadata := matches[3]
	isArray := matches[4] == "[]"

	var metadata [2]uint16
	if rawMetadata != "" {
		metadata = [2]uint16{}
		// only decimal types can have metadata
		if baseType != DecimalStr {
			return nil, fmt.Errorf("metadata is only allowed for decimal type")
		}

		parts := strings.Split(rawMetadata, ",")
		// can be either DECIMAL(10,5) or just DECIMAL
		if len(parts) != 2 && len(parts) != 0 {
			return nil, fmt.Errorf("invalid metadata format: %s", rawMetadata)
		}
		for i, part := range parts {
			num, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("invalid metadata value: %s", part)
			}
			metadata[i] = uint16(num)
		}
	}

	baseName, ok := typeAlias[baseType]
	if !ok {
		return nil, fmt.Errorf("unknown data type: %s", baseType)
	}

	dt := &DataType{
		Name:     baseName,
		Metadata: metadata,
		IsArray:  isArray,
	}

	return dt, dt.Clean()
}

// maps type names to their base names
var typeAlias = map[string]string{
	"string":  textStr,
	"text":    textStr,
	"int":     intStr,
	"integer": intStr,
	"bigint":  intStr,
	"int8":    intStr,
	"bool":    boolStr,
	"boolean": boolStr,
	"blob":    blobStr,
	"bytea":   blobStr,
	"uuid":    uuidStr,
	"decimal": DecimalStr,
	"numeric": DecimalStr,
}
