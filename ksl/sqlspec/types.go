package sqlspec

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"ksl/kslspec"
)

func TypeSpec(t kslspec.Type) (*kslspec.ConcreteType, error) {
	switch t := t.(type) {
	case *TimeType:
		s := &kslspec.ConcreteType{Type: timeAlias(t.T)}
		if p := t.Precision; p != nil && *p != DefaultTimePrecision {
			s.AddIntAttr("precision", *p)
		}
		return s, nil
	case *ArrayType:
		sp, err := TypeSpec(t.Type)
		if err != nil {
			return nil, err
		}
		sp.AddBoolAttr("array", true)
		return sp, nil

	case *BitType:
		sp := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if sp.Type == TypeBit && t.Size > 1 || sp.Type == TypeBitVar && t.Size > 0 {
			sp.AddIntAttr("size", int(t.Size))
		}
		return sp, nil
	case *BoolType:
		return &kslspec.ConcreteType{Type: strings.ToLower(t.T)}, nil
	case *BinaryType:
		sp := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if t.Size != nil && *t.Size > 0 {
			sp.AddIntAttr("size", int(*t.Size))
		}
		return sp, nil
	case *CurrencyType:
		return &kslspec.ConcreteType{Type: strings.ToLower(t.T)}, nil
	case *EnumType:
		return &kslspec.ConcreteType{Type: strings.ToLower(t.T)}, nil
	case *IntegerType:
		sp := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if t.Unsigned {
			sp.AddBoolAttr("unsigned", true)
		}
		return sp, nil
	case *IntervalType:
		s := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if p := t.Precision; p != nil && *p != DefaultTimePrecision {
			s.AddIntAttr("precision", *p)
		}
		if t.F != "" {
			s.AddStringAttr("interval", t.F)
		}
		return s, nil
	case *StringType:
		sp := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if t.Size > 0 {
			sp.AddIntAttr("size", int(t.Size))
		}
		return sp, nil
	case *FloatType:
		sp := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if t.Precision > 0 {
			sp.AddIntAttr("precision", int(t.Precision))
		}
		if t.Unsigned {
			sp.AddBoolAttr("unsigned", true)
		}
		return sp, nil
	case *DecimalType:
		sp := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if t.Precision > 0 {
			sp.AddIntAttr("precision", int(t.Precision))
		}
		if t.Scale > 0 {
			sp.AddIntAttr("scale", int(t.Scale))
		}
		if t.Unsigned {
			sp.AddBoolAttr("unsigned", true)
		}
		return sp, nil
	case *SerialType:
		sp := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if t.Precision > 0 {
			sp.AddIntAttr("precision", int(t.Precision))
		}
		if t.SequenceName != "" {
			sp.AddStringAttr("sequence", t.SequenceName)
		}
		return sp, nil
	case *JSONType:
		return &kslspec.ConcreteType{Type: strings.ToLower(t.T)}, nil
	case *UUIDType:
		return &kslspec.ConcreteType{Type: strings.ToLower(t.T)}, nil
	case *SpatialType:
		return &kslspec.ConcreteType{Type: strings.ToLower(t.T)}, nil
	case *NetworkType:
		sp := &kslspec.ConcreteType{Type: strings.ToLower(t.T)}
		if t.Size > 0 {
			sp.AddIntAttr("size", int(t.Size))
		}
		return sp, nil
	case *UserDefinedType:
		return &kslspec.ConcreteType{Type: strings.ToLower(t.T)}, nil
	case *XMLType:
		return &kslspec.ConcreteType{Type: strings.ToLower(t.T)}, nil
	case *kslspec.UnsupportedType:
		return nil, fmt.Errorf("postgres: unsupported type: %q", t.T)
	default:
		return nil, fmt.Errorf("unsupported type %T", t)
	}
}

func ConvertType(ct *kslspec.ConcreteType) (kslspec.Type, error) {
	var size, precision, scale int
	var unsigned bool
	var sequenceName, interval string
	var array bool

	for _, a := range ct.Attrs {
		switch a.Name {
		case "size":
			size, _ = a.Int()
		case "precision":
			precision, _ = a.Int()
		case "scale":
			scale, _ = a.Int()
		case "unsigned":
			unsigned, _ = a.Bool()
		case "sequence":
			sequenceName, _ = a.String()
		case "array":
			array, _ = a.Bool()
		case "interval":
			interval, _ = a.String()
		}
	}

	var typ kslspec.Type
	switch t := ct.Type; strings.ToLower(t) {
	case TypeBigInt, TypeInt8, TypeInt, TypeInteger, TypeInt4, TypeSmallInt, TypeInt2:
		typ = &IntegerType{T: t}
	case TypeBit, TypeBitVar:
		typ = &BitType{T: t, Size: int64(size)}
	case TypeBool, TypeBoolean:
		typ = &BoolType{T: t}
	case TypeBytea:
		typ = &BinaryType{T: t}
	case TypeCharacter, TypeChar, TypeCharVar, TypeVarChar, TypeText:
		typ = &StringType{T: t, Size: size}
	case TypeCIDR, TypeInet, TypeMACAddr, TypeMACAddr8:
		typ = &NetworkType{T: t}
	case TypeCircle, TypeLine, TypeLseg, TypeBox, TypePath, TypePolygon, TypePoint:
		typ = &SpatialType{T: t}
	case TypeDate:
		typ = &TimeType{T: t}
	case TypeTime, TypeTimeWOTZ, TypeTimeTZ, TypeTimeWTZ, TypeTimestamp, TypeTimestampTZ, TypeTimestampWTZ, TypeTimestampWOTZ:
		p := DefaultTimePrecision
		if precision != 0 {
			p = precision
		}
		typ = &TimeType{T: t, Precision: &p}
	case TypeInterval:
		p := DefaultTimePrecision
		if precision != 0 {
			p = precision
		}
		it := &IntervalType{T: t, Precision: &p}
		if interval != "" {
			f, ok := intervalField(interval)
			if !ok {
				return &kslspec.UnsupportedType{T: interval}, nil
			}
			it.F = f
		}
		typ = it
	case TypeReal, TypeDouble, TypeFloat, TypeFloat4, TypeFloat8:
		typ = &FloatType{T: t, Precision: precision, Unsigned: unsigned}
	case TypeJSON, TypeJSONB:
		typ = &JSONType{T: t}
	case TypeMoney:
		typ = &CurrencyType{T: t}
	case TypeDecimal, TypeNumeric:
		typ = &DecimalType{T: t, Precision: precision, Scale: scale, Unsigned: unsigned}
	case TypeSmallSerial, TypeSerial, TypeBigSerial, TypeSerial2, TypeSerial4, TypeSerial8:
		typ = &SerialType{T: t, Precision: precision, SequenceName: sequenceName}
	case TypeUUID:
		typ = &UUIDType{T: t}
	case TypeXML:
		typ = &XMLType{T: t}
	case TypeUserDefined:
		typ = &UserDefinedType{T: t}
	default:
		typ = &UserDefinedType{T: t}
	}

	if array {
		typ = &ArrayType{Type: typ, T: ct.Type + "[]"}
	}
	return typ, nil
}

func PrintType(typ *kslspec.ConcreteType) (string, error) {
	var args []string
	var mid, suffix string

	for key, arg := range typ.Attrs {
		if key == "unsigned" {
			b, err := arg.Bool()
			if err != nil {
				return "", err
			}
			if b {
				suffix += " unsigned"
			}
			continue
		}
		if key == "array" {
			b, err := arg.Bool()
			if err != nil {
				return "", err
			}
			if b {
				mid = "[]"
			}
			continue
		}
		switch v := arg.Value.(type) {
		case *kslspec.LiteralValue:
			args = append(args, v.Value)
		case *kslspec.ListValue:
			for _, li := range v.Values {
				lit, ok := li.(*kslspec.LiteralValue)
				if !ok {
					return "", fmt.Errorf("expecting literal value. got: %T", li)
				}
				uq, err := strconv.Unquote(lit.Value)
				if err != nil {
					return "", fmt.Errorf("expecting list items to be quoted strings: %w", err)
				}
				args = append(args, "'"+uq+"'")
			}
		default:
			return "", fmt.Errorf("unsupported type %T for PrintType", v)
		}
	}
	if len(args) > 0 {
		mid = "(" + strings.Join(args, ",") + ")"
	}
	return typ.Type + mid + suffix, nil
}

// FormatType converts schema type to its column form in the database.
// An error is returned if the type cannot be recognized.
func FormatType(t kslspec.Type) (string, error) {
	var f string
	switch t := t.(type) {
	case *ArrayType:
		f = strings.ToLower(t.T)
	case *BitType:
		f = strings.ToLower(t.T)
		if f == TypeBit && t.Size > 1 || f == TypeBitVar && t.Size > 0 {
			f = fmt.Sprintf("%s(%d)", f, t.Size)
		}
	case *BoolType:
		if f = strings.ToLower(t.T); f == TypeBool {
			f = TypeBoolean
		}
	case *BinaryType:
		f = strings.ToLower(t.T)
	case *CurrencyType:
		f = strings.ToLower(t.T)
	case *EnumType:
		if t.T == "" {
			return "", errors.New("postgres: missing enum type name")
		}
		f = t.T
	case *IntegerType:
		switch f = strings.ToLower(t.T); f {
		case TypeSmallInt, TypeInteger, TypeBigInt:
		case TypeInt2:
			f = TypeSmallInt
		case TypeInt, TypeInt4:
			f = TypeInteger
		case TypeInt8:
			f = TypeBigInt
		}
	case *IntervalType:
		f = strings.ToLower(t.T)
		if t.F != "" {
			f += " " + strings.ToLower(t.F)
		}
		if t.Precision != nil && *t.Precision != DefaultTimePrecision {
			f += fmt.Sprintf("(%d)", *t.Precision)
		}
	case *StringType:
		switch f = strings.ToLower(t.T); f {
		case TypeText:
		case TypeChar, TypeCharacter:
			n := t.Size
			if n == 0 {
				n = 1
			}
			f = fmt.Sprintf("%s(%d)", TypeCharacter, n)
		case TypeVarChar, TypeCharVar:
			f = TypeCharVar
			if t.Size != 0 {
				f = fmt.Sprintf("%s(%d)", TypeCharVar, t.Size)
			}
		default:
			return "", fmt.Errorf("postgres: unexpected string type: %q", t.T)
		}
	case *TimeType:
		f = timeAlias(t.T)
		if p := t.Precision; p != nil && *p != DefaultTimePrecision && strings.HasPrefix(f, "time") {
			f += fmt.Sprintf("(%d)", *p)
		}
	case *FloatType:
		switch f = strings.ToLower(t.T); f {
		case TypeFloat4:
			f = TypeReal
		case TypeFloat8:
			f = TypeDouble
		case TypeFloat:
			switch {
			case t.Precision > 0 && t.Precision <= 24:
				f = TypeReal
			case t.Precision == 0 || (t.Precision > 24 && t.Precision <= 53):
				f = TypeDouble
			default:
				return "", fmt.Errorf("postgres: precision for type float must be between 1 and 53: %d", t.Precision)
			}
		}
	case *DecimalType:
		switch f = strings.ToLower(t.T); f {
		case TypeNumeric:
		case TypeDecimal:
			f = TypeNumeric
		default:
			return "", fmt.Errorf("postgres: unexpected decimal type: %q", t.T)
		}
		switch p, s := t.Precision, t.Scale; {
		case p == 0 && s == 0:
		case s < 0:
			return "", fmt.Errorf("postgres: decimal type must have scale >= 0: %d", s)
		case p == 0 && s > 0:
			return "", fmt.Errorf("postgres: decimal type must have precision between 1 and 1000: %d", p)
		case s == 0:
			f = fmt.Sprintf("%s(%d)", f, p)
		default:
			f = fmt.Sprintf("%s(%d,%d)", f, p, s)
		}
	case *SerialType:
		switch f = strings.ToLower(t.T); f {
		case TypeSmallSerial, TypeSerial, TypeBigSerial:
		case TypeSerial2:
			f = TypeSmallSerial
		case TypeSerial4:
			f = TypeSerial
		case TypeSerial8:
			f = TypeBigSerial
		default:
			return "", fmt.Errorf("postgres: unexpected serial type: %q", t.T)
		}
	case *JSONType:
		f = strings.ToLower(t.T)
	case *UUIDType:
		f = strings.ToLower(t.T)
	case *SpatialType:
		f = strings.ToLower(t.T)
	case *NetworkType:
		f = strings.ToLower(t.T)
	case *UserDefinedType:
		f = strings.ToLower(t.T)
	case *XMLType:
		f = strings.ToLower(t.T)
	case *kslspec.UnsupportedType:
		return "", fmt.Errorf("postgres: unsupported type: %q", t.T)
	default:
		return "", fmt.Errorf("postgres: invalid schema type: %T", t)
	}
	return f, nil
}

// ParseType returns the Type value represented by the given raw type.
// The raw value is expected to follow the format in PostgreSQL information schema
// or as an input for the CREATE TABLE statement.
func ParseType(typ string) (kslspec.Type, error) {
	var (
		err error
		d   *columnDesc
	)
	// Normalize PostgreSQL array data types from "CREATE TABLE" format to
	// "INFORMATION_SCHEMA" format (i.e. as it is inspected from the database).
	if t, ok := arrayType(typ); ok {
		d = &columnDesc{typ: TypeArray, fmtype: t + "[]"}
	} else if d, err = parseColumn(typ); err != nil {
		return nil, err
	}
	t, err := columnType(d)
	if err != nil {
		return nil, err
	}
	// If the type is unknown (to us), we fall back to user-defined but expect
	// to improve this in future versions by ensuring this against the database.
	if ut, ok := t.(*kslspec.UnsupportedType); ok {
		t = &UserDefinedType{T: ut.T}
	}
	return t, nil
}

func columnType(c *columnDesc) (kslspec.Type, error) {
	var typ kslspec.Type
	switch t := c.typ; strings.ToLower(t) {
	case TypeBigInt, TypeInt8, TypeInt, TypeInteger, TypeInt4, TypeSmallInt, TypeInt2:
		typ = &IntegerType{T: t}
	case TypeBit, TypeBitVar:
		typ = &BitType{T: t, Size: c.size}
	case TypeBool, TypeBoolean:
		typ = &BoolType{T: t}
	case TypeBytea:
		typ = &BinaryType{T: t}
	case TypeCharacter, TypeChar, TypeCharVar, TypeVarChar, TypeText:
		typ = &StringType{T: t, Size: int(c.size)}
	case TypeCIDR, TypeInet, TypeMACAddr, TypeMACAddr8:
		typ = &NetworkType{T: t}
	case TypeCircle, TypeLine, TypeLseg, TypeBox, TypePath, TypePolygon, TypePoint:
		typ = &SpatialType{T: t}
	case TypeDate:
		typ = &TimeType{T: t}
	case TypeTime, TypeTimeWOTZ, TypeTimeTZ, TypeTimeWTZ, TypeTimestamp,
		TypeTimestampTZ, TypeTimestampWTZ, TypeTimestampWOTZ:
		p := DefaultTimePrecision
		if c.timePrecision != nil {
			p = int(*c.timePrecision)
		}
		typ = &TimeType{T: t, Precision: &p}
	case TypeInterval:
		p := DefaultTimePrecision
		if c.timePrecision != nil {
			p = int(*c.timePrecision)
		}
		typ = &IntervalType{T: t, Precision: &p}
		if c.interval != "" {
			f, ok := intervalField(c.interval)
			if !ok {
				return &kslspec.UnsupportedType{T: c.interval}, nil
			}
			typ.(*IntervalType).F = f
		}
	case TypeReal, TypeDouble, TypeFloat, TypeFloat4, TypeFloat8:
		typ = &FloatType{T: t, Precision: int(c.precision)}
	case TypeJSON, TypeJSONB:
		typ = &JSONType{T: t}
	case TypeMoney:
		typ = &CurrencyType{T: t}
	case TypeDecimal, TypeNumeric:
		typ = &DecimalType{T: t, Precision: int(c.precision), Scale: int(c.scale)}
	case TypeSmallSerial, TypeSerial, TypeBigSerial, TypeSerial2, TypeSerial4, TypeSerial8:
		typ = &SerialType{T: t, Precision: int(c.precision)}
	case TypeUUID:
		typ = &UUIDType{T: t}
	case TypeXML:
		typ = &XMLType{T: t}
	case TypeArray:
		typ = &ArrayType{T: c.fmtype}
		if t, ok := arrayType(c.fmtype); ok {
			tt, err := ParseType(t)
			if err != nil {
				return nil, err
			}
			if c.elemtyp == "e" {
				tt = newEnumType(t, c.typelem)
			}
			typ.(*ArrayType).Type = tt
		}
	case TypeUserDefined:
		typ = &UserDefinedType{T: c.fmtype}
		// The `typtype` column is set to 'e' for enum types, and the
		// values are filled in batch after the rows above is closed.
		// https://postgresql.org/docs/current/catalog-pg-type.html
		if c.typtype == "e" {
			typ = newEnumType(c.fmtype, c.typid)
		}
	default:
		typ = &kslspec.UnsupportedType{T: t}
	}
	return typ, nil
}

// reArray parses array declaration. See: https://postgresql.org/docs/current/arrays.html.
var reArray = regexp.MustCompile(`(?i)(.+?)(( +ARRAY( *\[[ \d]*] *)*)+|( *\[[ \d]*] *)+)$`)

// arrayType reports if the given string is an array type (e.g. int[], text[2]),
// and returns its "udt_name" as it was inspected from the database.
func arrayType(t string) (string, bool) {
	matches := reArray.FindStringSubmatch(t)
	if len(matches) < 2 {
		return "", false
	}
	return strings.TrimSpace(matches[1]), true
}

// reInterval parses declaration of interval fields. See: https://www.postgresql.org/docs/current/datatype-datetime.html.
var reInterval = regexp.MustCompile(`(?i)(?:INTERVAL\s*)?(YEAR|MONTH|DAY|HOUR|MINUTE|SECOND|YEAR TO MONTH|DAY TO HOUR|DAY TO MINUTE|DAY TO SECOND|HOUR TO MINUTE|HOUR TO SECOND|MINUTE TO SECOND)?\s*(?:\(([0-6])\))?$`)

// intervalField reports if the given string is an interval
// field type and returns its value (e.g. SECOND, MINUTE TO SECOND).
func intervalField(t string) (string, bool) {
	matches := reInterval.FindStringSubmatch(t)
	if len(matches) != 3 || matches[1] == "" {
		return "", false
	}
	return matches[1], true
}

// columnDesc represents a column descriptor.
type columnDesc struct {
	typ           string // data_type
	fmtype        string // pg_catalog.format_type
	size          int64  // character_maximum_length
	typtype       string // pg_type.typtype
	typelem       int64  // pg_type.typelem
	elemtyp       string // pg_type.typtype of the array element type above.
	typid         int64  // pg_type.oid
	precision     int64
	timePrecision *int64
	scale         int64
	parts         []string
	interval      string
}

var reDigits = regexp.MustCompile(`\d`)

func parseColumn(s string) (*columnDesc, error) {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '(' || r == ')' || r == ' ' || r == ','
	})
	var (
		err error
		c   = &columnDesc{
			typ:   parts[0],
			parts: parts,
		}
	)
	switch c.parts[0] {
	case TypeVarChar, TypeCharVar, TypeChar, TypeCharacter:
		if err := parseCharParts(c.parts, c); err != nil {
			return nil, err
		}
	case TypeDecimal, TypeNumeric, TypeFloat:
		if len(parts) > 1 {
			c.precision, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("postgres: parse precision %q: %w", parts[1], err)
			}
		}
		if len(parts) > 2 {
			c.scale, err = strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("postgres: parse scale %q: %w", parts[1], err)
			}
		}
	case TypeBit:
		if err := parseBitParts(parts, c); err != nil {
			return nil, err
		}
	case TypeDouble, TypeFloat8:
		c.precision = 53
	case TypeReal, TypeFloat4:
		c.precision = 24

	case TypeTime, TypeTimeTZ, TypeTimestamp, TypeTimestampTZ:
		t, p := s, int64(DefaultTimePrecision)
		// If the second part is only one digit it is the precision argument.
		// For cases like "timestamp(4) with time zone" make sure to not drop
		// the rest of the type definition.
		if len(parts) > 1 && reDigits.MatchString(parts[1]) {
			i, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("postgres: parse time precision %q: %w", parts[1], err)
			}
			p = i
			t = strings.Join(append(c.parts[:1], c.parts[2:]...), " ")
		}
		c.typ = timeAlias(t)
		c.timePrecision = &p
	case TypeInterval:
		matches := reInterval.FindStringSubmatch(s)
		c.interval = matches[1]
		if matches[2] != "" {
			i, err := strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("postgres: parse interval precision %q: %w", parts[1], err)
			}
			c.timePrecision = &i
		}
	default:
		c.typ = s
	}
	return c, nil
}

func parseCharParts(parts []string, c *columnDesc) error {
	j := strings.Join(parts, " ")
	switch {
	case strings.HasPrefix(j, TypeVarChar):
		c.typ = TypeVarChar
		parts = parts[1:]
	case strings.HasPrefix(j, TypeCharVar):
		c.typ = TypeCharVar
		parts = parts[2:]
	default:
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return nil
	}
	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("postgres: parse size %q: %w", parts[0], err)
	}
	c.size = size
	return nil
}

func parseBitParts(parts []string, c *columnDesc) error {
	if len(parts) == 1 {
		c.size = 1
		return nil
	}
	parts = parts[1:]
	if parts[0] == "varying" {
		c.typ = TypeBitVar
		parts = parts[1:]
	}
	if len(parts) == 0 {
		return nil
	}
	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("postgres: parse size %q: %w", parts[1], err)
	}
	c.size = size
	return nil
}

// timeAlias returns the abbreviation for the given time type.
func timeAlias(t string) string {
	switch t = strings.ToLower(t); t {
	// TIMESTAMPTZ be equivalent to TIMESTAMP WITH TIME ZONE.
	case TypeTimestampWTZ:
		t = TypeTimestampTZ
	// TIMESTAMP be equivalent to TIMESTAMP WITHOUT TIME ZONE.
	case TypeTimestampWOTZ:
		t = TypeTimestamp
	// TIME be equivalent to TIME WITHOUT TIME ZONE.
	case TypeTimeWOTZ:
		t = TypeTime
	// TIMETZ be equivalent to TIME WITH TIME ZONE.
	case TypeTimeWTZ:
		t = TypeTimeTZ
	}
	return t
}

var reEnumType = regexp.MustCompile(`^(?:(".+"|\w+)\.)?(".+"|\w+)$`)

func newEnumType(t string, id int64) *enumType {
	var (
		e     = &enumType{T: t, ID: id}
		parts = reEnumType.FindStringSubmatch(e.T)
		r     = func(s string) string {
			s = strings.ReplaceAll(s, `""`, `"`)
			if len(s) > 1 && s[0] == '"' && s[len(s)-1] == '"' {
				s = s[1 : len(s)-1]
			}
			return s
		}
	)
	if len(parts) > 1 {
		e.Schema = r(parts[1])
	}
	if len(parts) > 2 {
		e.T = r(parts[2])
	}
	return e
}

// IsUnique reports if the type is unique constraint.
func (c ConstraintType) IsUnique() bool { return strings.ToLower(c.Type) == "u" }

// IntegerType returns the underlying integer type this serial type represents.
func (s *SerialType) IntegerType() *IntegerType {
	t := &IntegerType{T: TypeInteger}
	switch s.T {
	case TypeSerial2, TypeSmallSerial:
		t.T = TypeSmallInt
	case TypeSerial8, TypeBigSerial:
		t.T = TypeBigInt
	}
	return t
}

// SetType sets the serial type from the given integer type.
func (s *SerialType) SetType(t *IntegerType) {
	switch t.T {
	case TypeSmallInt, TypeInt2:
		s.T = TypeSmallSerial
	case TypeInteger, TypeInt4, TypeInt:
		s.T = TypeSerial
	case TypeBigInt, TypeInt8:
		s.T = TypeBigSerial
	}
}

// sequence returns the inspected name of the sequence
// or the standard name defined by postgres.
func (s *SerialType) sequence(t *Table, c *Column) string {
	if s.SequenceName != "" {
		return s.SequenceName
	}
	return fmt.Sprintf("%s_%s_seq", t.Name, c.Name)
}

// newIndexStorage parses and returns the index storage parameters.
func newIndexStorage(opts string) (*IndexStorageParams, error) {
	params := &IndexStorageParams{}
	for _, p := range strings.Split(strings.Trim(opts, "{}"), ",") {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid index storage parameter: %s", p)
		}
		switch kv[0] {
		case "autosummarize":
			b, err := strconv.ParseBool(kv[1])
			if err != nil {
				return nil, fmt.Errorf("failed parsing autosummarize %q: %w", kv[1], err)
			}
			params.AutoSummarize = b
		case "pages_per_range":
			i, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed parsing pages_per_range %q: %w", kv[1], err)
			}
			params.PagesPerRange = i
		}
	}
	return params, nil
}
