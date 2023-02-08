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
	fs.String("server.addr", "", "the address of the Kwil-gateway server")
	fs.StringSlice("server.cors", []string{}, "the cors of the Kwil-gateway server, use comma to separate multiple cors")
	fs.String("server.healthcheck_key", "", "the health check api key of the Kwil-gateway server")
	fs.String("server.key_file", "", "the api key file of the Kwil-gateway server(default: $HOME/.kwilgw/keys.json)")

	// log flags
	fs.String("log.level", "", "the level of the log (default: config)")
	fs.StringSlice("log.output_paths", []string{}, "the output path of the log (default: ['stdout']), use comma to separate multiple output paths")

	// hasura flags
	fs.String("graphql.endpoint", "", "the endpoint of the Graphql server")

	// kwil flags
	fs.String("kwild.endpoint", "", "the endpoint of the Kwild server")
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
	viper.BindEnv("server.addr")
	viper.BindPFlag("server.addr", fs.Lookup("server.addr"))
	viper.SetDefault("server.addr", "0.0.0.0:8082")
	viper.BindEnv("server.cors")
	viper.BindPFlag("server.cors", fs.Lookup("server.cors"))
	viper.SetDefault("server.cors", []string{"*"})
	viper.BindEnv("server.healthcheck_key")
	viper.BindPFlag("server.healthcheck_key", fs.Lookup("server.healthcheck_key"))
	viper.SetDefault("server.healthcheck_key", "kwil-gateway-health-check-key")
	viper.BindEnv("server.key_file")
	viper.BindPFlag("server.key_file", fs.Lookup("server.key_file"))
	viper.SetDefault("server.key_file", filepath.Join(home, DefaultConfigDir, "keys.json"))

	// log key & env
	viper.BindEnv("log.level")
	viper.BindPFlag("log.level", fs.Lookup("log.level"))
	viper.SetDefault("log.level", "info")
	viper.BindEnv("log.output_paths")
	viper.BindPFlag("log.output_paths", fs.Lookup("log.output_paths"))
	viper.SetDefault("log.output_paths", []string{"stdout"})

	// hasura key & env
	viper.BindEnv("graphql.endpoint")
	viper.BindPFlag("graphql.endpoint", fs.Lookup("graphql.endpoint"))
	viper.SetDefault("graphql.endpoint", "http://localhost:8080")

	// kwil key & env
	viper.BindEnv("kwild.endpoint")
	viper.BindPFlag("kwild.endpoint", fs.Lookup("kwild.endpoint"))
	viper.SetDefault("kwild.endpoint", "localhost:50051")
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
