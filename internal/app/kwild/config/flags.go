package config

import (
	"github.com/kwilteam/kwil-db/pkg/config"

	"github.com/spf13/pflag"
)

func BindFlagsAndEnv(fs *pflag.FlagSet) {
	config.BindFlags(fs, RegisteredVariables)
}
