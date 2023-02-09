package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"kwil/internal/pkg/config"
	"kwil/pkg/fund"
	"kwil/pkg/log"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvPrefix         = "KWILd"
	DefaultConfigDir  = ".kwild"
	DefaultConfigName = "config"
	DefaultConfigType = "yaml"
)

type GraphqlConfig struct {
	Endpoint string `mapstructure:"endpoint"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Url      string `mapstructure:"url"`
	SslMode  string `mapstructure:"sslmode"`
}

func (c *PostgresConfig) DbUrl() string {
	if c.Url != "" {
		return c.Url
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SslMode)
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

type AppConfig struct {
	Server  ServerConfig   `mapstructure:"server"`
	Log     log.Config     `mapstructure:"log"`
	Fund    fund.Config    `mapstructure:"fund"`
	Db      PostgresConfig `mapstructure:"db"`
	Graphql GraphqlConfig  `mapstructure:"graphql"`
}

var ConfigFile string

// BindGlobalFlags binds the global flags to the command.
func BindGlobalFlags(fs *pflag.FlagSet) {
	// server flags
	fs.String("server.addr", "", "the address of the Kwil server")

	// log flags
	fs.String("log.level", "", "the level of the Kwil log (default: config)")
	fs.StringSlice("log.output_paths", []string{}, "the output path of the Kwil log (default: ['stdout']), use comma to separate multiple output paths")

	// db flags
	fs.String("db.host", "", "the host of the postgres database")
	fs.Int("db.port", 0, "the port of the postgres database")
	fs.String("db.username", "", "the username of the postgres database")
	fs.String("db.password", "", "the password of the postgres database")
	fs.String("db.database", "", "the database of the postgres database")
	fs.String("db.sslmode", "", "the sslmode of the postgres database")
	fs.String("db.url", "", "the url of the postgres database")

	// graphql flags
	fs.String("graphql.endpoint", "", "the endpoint of the Kwil graphql")

	// fund flags
	fs.String("fund.wallet", "", "you wallet private key")
	fs.String("fund.token_address", "", "the address of the funding pool token")
	fs.String("fund.pool_address", "", "the address of the funding pool")
	fs.String("fund.validator_address", "", "the address of the validator")
	fs.String("fund.chain_code", "", "the chain code of the funding pool chain")
	fs.String("fund.rpc_url", "", "the provider url of the funding pool chain")
	fs.Int64("fund.reconnect_interval", 0, "the reconnect interval of the funding pool")
	fs.Int64("fund.block_confirmation", 0, "the block confirmation of the funding pool")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(EnvPrefix)

	// server key & env
	viper.BindEnv("server.addr")
	viper.BindPFlag("server.addr", fs.Lookup("server.addr"))
	viper.SetDefault("server.addr", "0.0.0.0:50051")

	// log key & env
	viper.BindEnv("log.level")
	viper.BindPFlag("log.level", fs.Lookup("log.level"))
	viper.SetDefault("log.level", "info")
	viper.BindEnv("log.output_paths")
	viper.BindPFlag("log.output_paths", fs.Lookup("log.output_paths"))
	viper.SetDefault("log.output_paths", []string{"stdout"})

	// db key & env
	viper.BindEnv("db.host")
	viper.BindPFlag("db.host", fs.Lookup("db.host"))
	viper.BindEnv("db.port")
	viper.BindPFlag("db.port", fs.Lookup("db.port"))
	viper.BindEnv("db.username")
	viper.BindPFlag("db.username", fs.Lookup("db.username"))
	viper.BindEnv("db.password")
	viper.BindPFlag("db.password", fs.Lookup("db.password"))
	viper.BindEnv("db.database")
	viper.BindPFlag("db.database", fs.Lookup("db.database"))
	viper.BindEnv("db.sslmode")
	viper.BindPFlag("db.sslmode", fs.Lookup("db.sslmode"))
	viper.SetDefault("db.sslmode", "disable")
	viper.BindEnv("db.url")
	viper.BindPFlag("db.url", fs.Lookup("db.url"))

	// graphql key & env
	viper.BindEnv("graphql.endpoint")
	viper.BindPFlag("graphql.endpoint", fs.Lookup("graphql.endpoint"))

	// fund key & env
	viper.BindEnv("fund.wallet")
	viper.BindPFlag("fund.wallet", fs.Lookup("fund.wallet"))
	viper.BindEnv("fund.token_address")
	viper.BindPFlag("fund.token_address", fs.Lookup("fund.token_address"))
	viper.BindEnv("fund.pool_address")
	viper.BindPFlag("fund.pool_address", fs.Lookup("fund.pool_address"))
	viper.BindEnv("fund.validator_address")
	viper.BindPFlag("fund.validator_address", fs.Lookup("fund.validator_address"))
	viper.BindEnv("fund.chain_code")
	viper.BindPFlag("fund.chain_code", fs.Lookup("fund.chain_code"))
	viper.BindEnv("fund.rpc_url")
	viper.BindPFlag("fund.rpc_url", fs.Lookup("fund.rpc_url"))
	viper.BindEnv("fund.reconnect_interval")
	viper.BindPFlag("fund.reconnect_interval", fs.Lookup("fund.reconnect_interval"))
	viper.SetDefault("fund.reconnect_interval", 30)
	viper.BindEnv("fund.block_confirmation")
	viper.BindPFlag("fund.block_confirmation", fs.Lookup("fund.block_confirmation"))
	viper.SetDefault("fund.block_confirmation", 12)
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

	if err = viper.Unmarshal(&cfg, viper.DecodeHook(config.StringPrivateKeyHookFunc())); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshal config file:", err)
	}
	return cfg, nil
}
