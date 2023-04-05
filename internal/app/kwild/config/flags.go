package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func BindGlobalFlags(fs *pflag.FlagSet) {
	// top-level config
	fs.String(PrivateKey.EnvName, "", "Private key of the node")
	fs.Int(GrpcListenAddress.EnvName, 0, "Port of the node")
	fs.String(SqliteFilePath.EnvName, "", "Sqlite file path of the node")
	fs.String(HttpListenAddress.EnvName, "", "Http listen address of the node")

	// Deposits
	fs.Int(DepositsReconnectionInterval.EnvName, 0, "Reconnection interval of the deposits")
	fs.Int(DepositsBlockConfirmation.EnvName, 0, "Block confirmation of the deposits")
	fs.Int(DepositsChainCode.EnvName, 0, "Chain code of the deposits")
	fs.String(DepositsClientChainRPCURL.EnvName, "", "Client chain RPC URL of the deposits")

	// Log
	fs.String(LogLevel.EnvName, "", "Log level of the node")
	fs.String(LogOutputPaths.EnvName, "", "Log output paths of the node")
}

func BindGlobalEnv(fs *pflag.FlagSet) {
	for _, v := range RegisteredVariables {
		viper.BindEnv(v.EnvName)
		viper.BindPFlag(v.EnvName, fs.Lookup(v.EnvName))
	}
}
