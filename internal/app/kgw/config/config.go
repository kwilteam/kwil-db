package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"kwil/internal/pkg/config"
	"kwil/pkg/log"
	"os"
)

const (
	EnvPrefix         = "KWILGW"
	DefaultConfigDir  = ".kwilgw"
	DefaultConfigName = "config"
	DefaultConfigType = "yaml"
)

// viper keys
const (
	ServerListenAddrKey     = "server.listen_addr"
	ServerCorsKey           = "server.cors"
	ServerHealthcheckKeyKey = "server.healthcheck_key"

	LogLevelKey       = "log.level"
	LogOutputPathsKey = "log.output_paths"

	GraphqlAddrKey = "graphql.addr"

	KwildAddrKey = "kwild.addr"
)

type GraphqlConfig struct {
	Addr string `mapstructure:"addr"`
}

type KwildConfig struct {
	Addr string `mapstructure:"addr"`
}

type ServerConfig struct {
	ListenAddr     string   `mapstructure:"listen_addr"`
	Cors           []string `mapstructure:"cors"`
	HealthcheckKey string   `mapstructure:"healthcheck_key"`
}

type AppConfig struct {
	Server  ServerConfig  `mapstructure:"server"`
	Log     log.Config    `mapstructure:"log"`
	Graphql GraphqlConfig `mapstructure:"graphql"`
	Kwild   KwildConfig   `mapstructure:"kwild"`
}

var ConfigFile string

var defaultConfig = map[string]interface{}{
	"server": map[string]interface{}{
		"listen_addr":     "0.0.0.0:8082",
		"cors":            []string{"*"},
		"healthcheck_key": "kwil-gateway-healthcheck-key",
	},
	"log": map[string]interface{}{
		"level":        "info",
		"output_paths": []string{"stdout"},
	},
	"graphql": map[string]interface{}{
		"addr": "http://localhost:8080",
	},
	"kwild": map[string]interface{}{
		"addr": "localhost:50051",
	},
}

// BindGlobalFlags binds the global flags to the command.
func BindGlobalFlags(fs *pflag.FlagSet) {
	// server flags
	fs.String(ServerListenAddrKey, "", "the address of the Kwil-gateway server")
	fs.StringSlice(ServerCorsKey, []string{}, "the cors of the Kwil-gateway server, use comma to separate multiple cors")
	fs.String(ServerHealthcheckKeyKey, "", "the health check api key of the Kwil-gateway server")

	// log flags
	fs.String(LogLevelKey, "", "the level of the log (default: info)")
	fs.StringSlice(LogOutputPathsKey, []string{}, "the output path of the log (default: ['stdout']), use comma to separate multiple output paths")

	// hasura flags
	fs.String(GraphqlAddrKey, "", "the address of the Graphql server")

	// kwil flags
	fs.String(KwildAddrKey, "", "the address of the Kwild server")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(EnvPrefix)

	envs := []string{
		GraphqlAddrKey,
		KwildAddrKey,
		LogLevelKey,
		LogOutputPathsKey,
		ServerListenAddrKey,
		ServerCorsKey,
		ServerHealthcheckKeyKey,
	}

	for _, v := range envs {
		viper.BindEnv(v)
		viper.BindPFlag(v, fs.Lookup(v))
	}
}

func LoadConfig() (cfg *AppConfig, err error) {
	config.LoadConfig(defaultConfig, ConfigFile, EnvPrefix, DefaultConfigName, DefaultConfigDir, DefaultConfigType)

	if err = viper.Unmarshal(&cfg, viper.DecodeHook(config.StringPrivateKeyHookFunc())); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshal config file:", err)
	}
	return cfg, nil
}
