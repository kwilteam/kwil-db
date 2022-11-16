package ksl

type Type interface{ typ() }

type builtin struct {
	Int      BuiltInScalar
	BigInt   BuiltInScalar
	Float    BuiltInScalar
	String   BuiltInScalar
	Bool     BuiltInScalar
	Date     BuiltInScalar
	DateTime BuiltInScalar
	Time     BuiltInScalar
	Bytes    BuiltInScalar
	Decimal  BuiltInScalar
}

func (builtin) From(name string) (BuiltInScalar, bool) {
	switch name {
	case "int", "bigint", "float", "string", "bool", "date", "datetime", "time", "bytes", "decimal":
		return BuiltInScalar{name}, true
	default:
		return BuiltInScalar{}, false
	}
}

type BuiltInScalar struct{ name string }

func (t BuiltInScalar) Name() string   { return t.name }
func (t BuiltInScalar) String() string { return t.name }

var BuiltIns = builtin{
	Int:      BuiltInScalar{"int"},
	BigInt:   BuiltInScalar{"bigint"},
	Float:    BuiltInScalar{"float"},
	String:   BuiltInScalar{"string"},
	Bool:     BuiltInScalar{"bool"},
	Date:     BuiltInScalar{"date"},
	DateTime: BuiltInScalar{"datetime"},
	Time:     BuiltInScalar{"time"},
	Bytes:    BuiltInScalar{"bytes"},
	Decimal:  BuiltInScalar{"decimal"},
}

type NativeType struct {
	Name string
	Args []string
}

type UnsupportedType struct {
	Name string
}

type UserDefinedType struct {
	Name string
}

func (BuiltInScalar) typ()   {}
func (NativeType) typ()      {}
func (UnsupportedType) typ() {}
func (UserDefinedType) typ() {}
