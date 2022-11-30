package postgres

import (
	"fmt"
	"ksl"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type PostgresType interface {
	ksl.Type
	pgtyp()
	String() string
	Name() string
	Args() []string
}

type EnumType struct {
	Name   string
	Values []string
}

var Types pgt

type pgt struct{}

func (pgt) Numeric(args ...int) PostgresType {
	return lo.Must(Types.ScalarFrom(TypeDecimal, itos(args)...))
}
func (pgt) Char(args ...int) PostgresType { return lo.Must(Types.ScalarFrom(TypeChar, itos(args)...)) }
func (pgt) VarChar(args ...int) PostgresType {
	return lo.Must(Types.ScalarFrom(TypeVarChar, itos(args)...))
}
func (pgt) Time(args ...int) PostgresType { return lo.Must(Types.ScalarFrom(TypeTime, itos(args)...)) }
func (pgt) TimeTZ(args ...int) PostgresType {
	return lo.Must(Types.ScalarFrom(TypeTimeTZ, itos(args)...))
}
func (pgt) Bit(args ...int) PostgresType { return lo.Must(Types.ScalarFrom(TypeBit, itos(args)...)) }
func (pgt) VarBit(args ...int) PostgresType {
	return lo.Must(Types.ScalarFrom(TypeVarBit, itos(args)...))
}

func (pgt) Timestamp(args ...int) PostgresType {
	return lo.Must(Types.ScalarFrom(TypeTimestamp, itos(args)...))
}
func (pgt) TimestampTZ(args ...int) PostgresType {
	return lo.Must(Types.ScalarFrom(TypeTimestampTZ, itos(args)...))
}

func (pgt) SmallInt() PostgresType { return scalartype{name: TypeSmallInt} }
func (pgt) Integer() PostgresType  { return scalartype{name: TypeInteger} }
func (pgt) BigInt() PostgresType   { return scalartype{name: TypeBigInt} }
func (pgt) Money() PostgresType    { return scalartype{name: TypeMoney} }
func (pgt) Network() PostgresType  { return scalartype{name: TypeInet} }
func (pgt) CIText() PostgresType   { return scalartype{name: TypeCIText} }
func (pgt) Real() PostgresType     { return scalartype{name: TypeReal} }
func (pgt) Double() PostgresType   { return scalartype{name: TypeDouble} }
func (pgt) Text() PostgresType     { return scalartype{name: TypeText} }
func (pgt) Date() PostgresType     { return scalartype{name: TypeDate} }
func (pgt) Bytes() PostgresType    { return scalartype{name: TypeByteA} }
func (pgt) Boolean() PostgresType  { return scalartype{name: TypeBoolean} }
func (pgt) UUID() PostgresType     { return scalartype{name: TypeUUID} }
func (pgt) XML() PostgresType      { return scalartype{name: TypeXML} }
func (pgt) Json() PostgresType     { return scalartype{name: TypeJson} }
func (pgt) JsonB() PostgresType    { return scalartype{name: TypeJsonB} }

func (pgt) IsNativeScalar(name string) bool {
	switch strings.ToLower(name) {
	case TypeBit, TypeBitVar, TypeVarBit, TypeBoolean, TypeBool, TypeByteA, TypeCharacter, TypeChar, TypeCharVar, TypeVarChar,
		TypeText, TypeCIText, TypeSmallInt, TypeInteger, TypeBigInt, TypeInt, TypeInt2, TypeInt4, TypeInt8, TypeCIDR, TypeInet,
		TypeMACAddr, TypeMACAddr8, TypeDate, TypeTime, TypeTimeTZ, TypeTimeWTZ, TypeTimeWOTZ, TypeTimestamp, TypeTimestampTZ,
		TypeTimestampWTZ, TypeTimestampWOTZ, TypeDouble, TypeReal, TypeFloat8, TypeFloat4, TypeFloat, TypeNumeric, TypeDecimal,
		TypeSmallSerial, TypeSerial, TypeBigSerial, TypeSerial2, TypeSerial4, TypeSerial8, TypeXML, TypeJson, TypeJsonB, TypeUUID,
		TypeMoney, TypeInterval:
		return true
	default:
		return false
	}
}

func (pgt) ScalarFrom(name string, params ...string) (PostgresType, error) {
	name = AliasType(name)
	var err error

	switch name {
	case TypeNumeric:
		if len(params) == 0 {
			break
		}
		if len(params) > 2 {
			return nil, fmt.Errorf("%q takes at most two arguments", name)
		}

		var precision, scale int
		if len(params) > 0 {
			if precision, err = strconv.Atoi(params[0]); err != nil {
				return nil, fmt.Errorf("invalid precision: %w", err)
			}
		}

		if precision < 1 || precision > 1000 {
			return nil, fmt.Errorf("precision must be between 1 and 1000")
		}

		if len(params) > 1 {
			if scale, err = strconv.Atoi(params[1]); err != nil {
				return nil, fmt.Errorf("invalid scale: %w", err)
			}
		}
		if scale < 0 || scale > 1000 {
			return nil, fmt.Errorf("scale must be between 0 and 1000")
		}

	case TypeBit, TypeVarBit, TypeChar, TypeVarChar:
		if len(params) == 0 {
			break
		}

		if len(params) > 1 {
			return nil, fmt.Errorf("%q takes at most one argument", name)
		}

		val, err := strconv.Atoi(params[0])
		if err != nil {
			return nil, err
		}

		if val == 0 {
			return nil, fmt.Errorf("length must be a positive integer")
		}

	case TypeTimestamp, TypeTimestampTZ, TypeTime, TypeTimeTZ, TypeInterval:
		if len(params) == 0 {
			break
		}

		if len(params) > 1 {
			return nil, fmt.Errorf("%q takes at most one argument", name)
		}

		val, err := strconv.Atoi(params[0])
		if err != nil {
			return nil, err
		}

		if val < 0 || val > 6 {
			return nil, fmt.Errorf("fractional seconds precision must be between 0 and 6")
		}
	default:
		if !Types.IsNativeScalar(name) {
			return UnsupportedType{name: name, args: params}, nil
		}
	}

	return scalartype{name: name, args: params}, nil
}

type UnsupportedType struct {
	ksl.Type
	name string
	args []string
}

func (t UnsupportedType) Name() string   { return t.name }
func (t UnsupportedType) Args() []string { return nil }
func (t UnsupportedType) String() string { return t.name }
func (UnsupportedType) pgtyp()           {}

type UserDefinedType struct {
	ksl.Type
	name string
}

func (t UserDefinedType) Name() string   { return t.name }
func (t UserDefinedType) Args() []string { return nil }
func (t UserDefinedType) String() string { return t.name }
func (UserDefinedType) pgtyp()           {}

type scalartype struct {
	ksl.Type
	name string
	args []string
}

func (scalartype) pgtyp() {}

func (t scalartype) String() string {
	if len(t.args) == 0 {
		return strings.ToUpper(t.name)
	}
	return strings.ToUpper(t.name) + "(" + strings.Join(t.args, ", ") + ")"
}
func (t scalartype) Name() string   { return t.name }
func (t scalartype) Args() []string { return t.args[:] }

func CompatibleScalar(name string) ksl.BuiltInScalar {
	switch AliasType(name) {
	case TypeSmallInt, TypeInteger, TypeInt:
		return ksl.BuiltIns.Int
	case TypeSerial, TypeSmallSerial:
		return ksl.BuiltIns.Int
	case TypeBigInt:
		return ksl.BuiltIns.BigInt
	case TypeBigSerial:
		return ksl.BuiltIns.BigInt
	case TypeDecimal, TypeMoney, TypeNumeric:
		return ksl.BuiltIns.Decimal
	case TypeReal, TypeDouble, TypeFloat:
		return ksl.BuiltIns.Float
	case TypeByteA:
		return ksl.BuiltIns.Bytes
	case TypeTimestamp, TypeTimestampTZ:
		return ksl.BuiltIns.DateTime
	case TypeDate:
		return ksl.BuiltIns.Date
	case TypeTime, TypeTimeTZ:
		return ksl.BuiltIns.Time
	case TypeBoolean, TypeBool:
		return ksl.BuiltIns.Bool
	default:
		return ksl.BuiltIns.String
	}
}

func AliasType(name string) string {
	switch name := strings.ToLower(name); name {
	case TypeBitVar:
		return TypeVarBit
	case TypeInt2:
		return TypeSmallInt
	case TypeInt4, TypeInt:
		return TypeInteger
	case TypeInt8:
		return TypeBigInt
	case TypeSerial2:
		return TypeSmallSerial
	case TypeSerial4:
		return TypeSerial
	case TypeSerial8:
		return TypeBigSerial
	case TypeFloat4:
		return TypeReal
	case TypeFloat8:
		return TypeDouble
	case TypeDecimal:
		return TypeNumeric
	case TypeCharVar:
		return TypeVarChar
	case TypeCharacter:
		return TypeChar
	case TypeTimeWOTZ:
		return TypeTime
	case TypeTimeWTZ:
		return TypeTimeTZ
	case TypeTimestampWOTZ:
		return TypeTimestamp
	case TypeTimestampWTZ:
		return TypeTimestampTZ
	case TypeBool:
		return TypeBoolean
	default:
		return name
	}
}

func ParseNativeType(name string, args ...string) (ksl.Type, error) {
	return Types.ScalarFrom(name, args...)
}

func ScalarTypeForNativeType(t ksl.Type) ksl.BuiltInScalar {
	switch t := t.(type) {
	case PostgresType:
		return CompatibleScalar(t.Name())
	case ksl.BuiltInScalar:
		return t
	default:
		panic(fmt.Errorf("unexpected type %T", t))
	}
}

func DefaultNativeTypeForScalar(t ksl.BuiltInScalar) PostgresType {
	if n, ok := scalarTypeDefaults[t]; ok {
		return n
	}
	return Types.Text()
}

func NativeTypeAnnotations() []string {
	return nativeTypeAnnots[:]
}

var nativeTypeAnnots = [...]string{
	TypeBit, TypeVarBit, TypeByteA, TypeChar, TypeVarChar, TypeTimestampTZ, TypeTimeTZ, TypeUUID,
	TypeCIText, TypeCIDR, TypeInet, TypeMACAddr, TypeMACAddr8, TypeXML, TypeMoney, TypeJson, TypeJsonB,
}

var scalarTypeDefaults = map[ksl.Type]PostgresType{
	ksl.BuiltIns.Int:      Types.Integer(),
	ksl.BuiltIns.BigInt:   Types.BigInt(),
	ksl.BuiltIns.Float:    Types.Double(),
	ksl.BuiltIns.Decimal:  Types.Numeric(65, 30),
	ksl.BuiltIns.Bool:     Types.Boolean(),
	ksl.BuiltIns.String:   Types.Text(),
	ksl.BuiltIns.DateTime: Types.Timestamp(3),
	ksl.BuiltIns.Date:     Types.Date(),
	ksl.BuiltIns.Time:     Types.Time(3),
	ksl.BuiltIns.Bytes:    Types.Bytes(),
}
