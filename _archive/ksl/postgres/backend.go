package postgres

import (
	"ksl"
	"ksl/backend"
	"ksl/sqlmigrate"
	"ksl/sqlschema"
)

func init() {
	backend.Register(Backend{})
	backend.RegisterDefault(Backend{})
}

type Backend struct{}

func (Backend) Name() string              { return "postgres" }
func (Backend) MaxIdentifierLength() uint { return 63 }

func (Backend) ScalarTypeForNativeType(t ksl.Type) ksl.BuiltInScalar {
	return ScalarTypeForNativeType(t)
}

func (Backend) DefaultNativeTypeForScalar(t ksl.BuiltInScalar) ksl.Type {
	return DefaultNativeTypeForScalar(t)
}

func (Backend) ParseNativeType(name string, args ...string) (ksl.Type, error) {
	return Types.ScalarFrom(name, args...)
}

func (Backend) ColumnTypeChange(old, new sqlschema.ColumnWalker) sqlmigrate.ColumnTypeChange {
	return columnTypeChange(old, new)
}
