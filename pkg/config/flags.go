package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type FlagType string

const (
	FlagTypeString  FlagType = "string"
	FlagTypeInt     FlagType = "int"
	FlagTypeInt64   FlagType = "int64"
	FlagTypeInt32   FlagType = "int32"
	FlagTypeUint    FlagType = "uint"
	FlagTypeUint64  FlagType = "uint64"
	FlagTypeUint32  FlagType = "uint32"
	FlagTypeBool    FlagType = "bool"
	FlagTypeFloat   FlagType = "float"
	FlagTypeFloat64 FlagType = "float64"
	FlagTypeFloat32 FlagType = "float32"
)

type Flag struct {
	Name  string
	Type  FlagType
	Usage string
}

func (f *Flag) IsEmpty() bool {
	return f.Name == ""
}

func (f *Flag) Bind(fs *pflag.FlagSet) {
	switch f.Type {
	case FlagTypeString:
		fs.String(f.Name, "", f.Usage)
	case FlagTypeInt:
		fs.Int(f.Name, 0, f.Usage)
	case FlagTypeInt64:
		fs.Int64(f.Name, 0, f.Usage)
	case FlagTypeInt32:
		fs.Int32(f.Name, 0, f.Usage)
	case FlagTypeUint:
		fs.Uint(f.Name, 0, f.Usage)
	case FlagTypeUint64:
		fs.Uint64(f.Name, 0, f.Usage)
	case FlagTypeUint32:
		fs.Uint32(f.Name, 0, f.Usage)
	case FlagTypeBool:
		fs.Bool(f.Name, false, f.Usage)
	case FlagTypeFloat:
		fs.Float64(f.Name, 0, f.Usage)
	case FlagTypeFloat64:
		fs.Float64(f.Name, 0, f.Usage)
	case FlagTypeFloat32:
		fs.Float32(f.Name, 0, f.Usage)
	default:
		panic("unknown flag type")
	}
}

// BindFlags binds all registered variables to the given flag set
func BindFlags(fs *pflag.FlagSet, RegisteredVariables []CfgVar) {
	for _, v := range RegisteredVariables {
		viper.BindEnv(v.EnvName)

		if !v.Flag.IsEmpty() {
			v.Flag.Bind(fs)
			viper.BindPFlag(v.EnvName, fs.Lookup(v.Flag.Name))
		}
	}
}
