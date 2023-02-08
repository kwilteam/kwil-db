package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"kwil/pkg/log"
	"kwil/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvPrefix         = "KWILGW"
	DefaultConfigDir  = ".kwilgw"
	DefaultConfigName = "config"
	DefaultConfigType = "yaml"
)

// viper keys
const (
	ServerAddrKey           = "server.addr"
	ServerCorsKey           = "server.cors"
	ServerHealthcheckKeyKey = "server.healthcheck_key"
	ServerKeyFileKey        = "server.key_file"

	LogLevelKey       = "log.level"
	LogOutputPathsKey = "log.output_paths"

	GraphqlEndpointKey = "graphql.endpoint"

	KwildEndpointKey = "kwild.endpoint"
)

type GraphqlConfig struct {
	Endpoint string `mapstructure:"endpoint"`
}

type KwildConfig struct {
	Endpoint string `mapstructure:"endpoint"`
}

type ServerConfig struct {
	Addr           string   `mapstructure:"addr"`
	Cors           []string `mapstructure:"cors"`
	HealthcheckKey string   `mapstructure:"healthcheck_key"`
	KeyFile        string   `mapstructure:"key_file"`
}

type AppConfig struct {
	Server  ServerConfig  `mapstructure:"server"`
	Log     log.Config    `mapstructure:"log"`
	Graphql GraphqlConfig `mapstructure:"graphql"`
	Kwild   KwildConfig   `mapstructure:"kwild"`
}

var ConfigFile string

// BindGlobalFlags binds the global flags to the command.
func BindGlobalFlags(fs *pflag.FlagSet) {
	// server flags
	fs.String(ServerAddrKey, "", "the address of the Kwil-gateway server")
	fs.StringSlice(ServerCorsKey, []string{}, "the cors of the Kwil-gateway server, use comma to separate multiple cors")
	fs.String(ServerHealthcheckKeyKey, "", "the health check api key of the Kwil-gateway server")
	fs.String(ServerKeyFileKey, "", "the api key file of the Kwil-gateway server(default: $HOME/.kwilgw/keys.json)")

	// log flags
	fs.String(LogLevelKey, "", "the level of the log (default: config)")
	fs.StringSlice(LogOutputPathsKey, []string{}, "the output path of the log (default: ['stdout']), use comma to separate multiple output paths")

	// hasura flags
	fs.String("GraphqlEndpointKey", "", "the endpoint of the Graphql server")

	// kwil flags
	fs.String(KwildEndpointKey, "", "the endpoint of the Kwild server")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(EnvPrefix)
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// server key & env
	viper.BindEnv(ServerAddrKey)
	viper.BindPFlag(ServerAddrKey, fs.Lookup(ServerAddrKey))
	viper.SetDefault(ServerAddrKey, "0.0.0.0:8082")
	viper.BindEnv(ServerCorsKey)
	viper.BindPFlag(ServerCorsKey, fs.Lookup(ServerCorsKey))
	viper.SetDefault(ServerCorsKey, []string{"*"})
	viper.BindEnv(ServerHealthcheckKeyKey)
	viper.BindPFlag(ServerHealthcheckKeyKey, fs.Lookup(ServerHealthcheckKeyKey))
	viper.SetDefault(ServerHealthcheckKeyKey, "kwil-gateway-health-check-key")
	viper.BindEnv(ServerKeyFileKey)
	viper.BindPFlag(ServerKeyFileKey, fs.Lookup(ServerKeyFileKey))
	viper.SetDefault(ServerKeyFileKey, filepath.Join(home, DefaultConfigDir, "keys.json"))

	// log key & env
	viper.BindEnv(LogLevelKey)
	viper.BindPFlag(LogLevelKey, fs.Lookup(LogLevelKey))
	viper.SetDefault(LogLevelKey, "info")
	viper.BindEnv(LogOutputPathsKey)
	viper.BindPFlag(LogOutputPathsKey, fs.Lookup(LogOutputPathsKey))
	viper.SetDefault(LogOutputPathsKey, []string{"stdout"})

	// hasura key & env
	viper.BindEnv(GraphqlEndpointKey)
	viper.BindPFlag(GraphqlEndpointKey, fs.Lookup(GraphqlEndpointKey))
	viper.SetDefault(GraphqlEndpointKey, "http://localhost:8080")

	// kwil key & env
	viper.BindEnv(KwildEndpointKey)
	viper.BindPFlag(KwildEndpointKey, fs.Lookup(KwildEndpointKey))
	viper.SetDefault(KwildEndpointKey, "localhost:50051")
}

func LoadConfig() (cfg *AppConfig, err error) {
	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
		fmt.Fprintln(os.Stdout, "using config file:", viper.ConfigFileUsed())
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		viper.AddConfigPath(filepath.Join(home, DefaultConfigDir))
		viper.SetConfigName(DefaultConfigName)
		viper.SetConfigType(DefaultConfigType)

		viper.SafeWriteConfig()
	}

	// PREFIX_A_B will be mapped to a.b
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(EnvPrefix)

	//viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	//viper.Debug()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// cfg file not found; ignore error if desired
			fmt.Fprintln(os.Stdout, "config file not found")
		} else {
			// cfg file was found but another error was produced
			return nil, fmt.Errorf("rrror loading config file: %s", err)
		}
	}

	if err = viper.Unmarshal(&cfg, viper.DecodeHook(utils.StringPrivateKeyHookFunc())); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshal config file:", err)
	}
	return cfg, nil
}
