package config

import (
	"kwil/pkg/config"

	"github.com/spf13/pflag"
)

func BindFlagsAndEnv(fs *pflag.FlagSet) {
	config.BindFlags(fs, RegisteredVariables)
}
