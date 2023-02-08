package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"kwil/pkg/fund"
	"kwil/pkg/log"
	"kwil/pkg/utils"
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

// viper keys
const (
	ServerAddrKey = "server.addr"

	LogLevelKey       = "log.level"
	LogOutputPathsKey = "log.output_paths"

	DBHostKey     = "db.host"
	DBPortKey     = "db.port"
	DBUsernameKey = "db.username"
	DBPasswordKey = "db.password"
	DBDatabaseKey = "db.database"
	DBUrlKey      = "db.url"
	DBSslModeKey  = "db.sslmode"

	GraphqlEndpointKey = "graphql.endpoint"

	FundWalletKey            = "fund.wallet"
	FundPoolAddressKey       = "fund.pool_address"
	FundChainCodeKey         = "fund.chain_code"
	FundRPCURLKey            = "fund.rpc_url"
	FundReconnectIntervalKey = "fund.reconnect_interval"
	FundBlockConfirmationKey = "fund.block_confirmation"
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
	DB      PostgresConfig `mapstructure:"db"`
	Graphql GraphqlConfig  `mapstructure:"graphql"`
}

var ConfigFile string

// BindGlobalFlags binds the global flags to the command.
func BindGlobalFlags(fs *pflag.FlagSet) {
	// server flags
	fs.String(ServerAddrKey, "", "the address of the Kwil server")

	// log flags
	fs.String(LogLevelKey, "", "the level of the Kwil log (default: config)")
	fs.StringSlice(LogOutputPathsKey, []string{}, "the output path of the Kwil log (default: ['stdout']), use comma to separate multiple output paths")

	// db flags
	fs.String(DBHostKey, "", "the host of the postgres database")
	fs.Int(DBPortKey, 0, "the port of the postgres database")
	fs.String(DBUsernameKey, "", "the username of the postgres database")
	fs.String(DBPasswordKey, "", "the password of the postgres database")
	fs.String(DBDatabaseKey, "", "the database of the postgres database")
	fs.String(DBSslModeKey, "", "the sslmode of the postgres database")
	fs.String(DBUrlKey, "", "the url of the postgres database(if set, the other db flags will be ignored)")

	// graphql flags
	fs.String(GraphqlEndpointKey, "", "the endpoint of the Kwil graphql")

	// fund flags
	fs.String(FundWalletKey, "", "you wallet private key")
	fs.String(FundPoolAddressKey, "", "the address of the funding pool")
	fs.String(FundChainCodeKey, "", "the chain code of the funding pool chain")
	fs.String(FundRPCURLKey, "", "the provider url of the funding pool chain")
	fs.Int64(FundReconnectIntervalKey, 0, "the reconnect interval of the funding pool")
	fs.Int64(FundBlockConfirmationKey, 0, "the block confirmation of the funding pool")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(EnvPrefix)

	// server key & env
	viper.BindEnv(ServerAddrKey)
	viper.BindPFlag(ServerAddrKey, fs.Lookup(ServerAddrKey))
	viper.SetDefault(ServerAddrKey, "0.0.0.0:50051")

	// log key & env
	viper.BindEnv(LogLevelKey)
	viper.BindPFlag(LogLevelKey, fs.Lookup(LogLevelKey))
	viper.SetDefault(LogLevelKey, "info")
	viper.BindEnv(LogOutputPathsKey)
	viper.BindPFlag(LogOutputPathsKey, fs.Lookup(LogOutputPathsKey))
	viper.SetDefault(LogOutputPathsKey, []string{"stdout"})

	// db key & env
	viper.BindEnv(DBHostKey)
	viper.BindPFlag(DBHostKey, fs.Lookup(DBHostKey))
	viper.BindEnv(DBPortKey)
	viper.BindPFlag(DBPortKey, fs.Lookup(DBPortKey))
	viper.BindEnv(DBUsernameKey)
	viper.BindPFlag(DBUsernameKey, fs.Lookup(DBUsernameKey))
	viper.BindEnv(DBPasswordKey)
	viper.BindPFlag(DBPasswordKey, fs.Lookup(DBPasswordKey))
	viper.BindEnv(DBDatabaseKey)
	viper.BindPFlag(DBDatabaseKey, fs.Lookup(DBDatabaseKey))
	viper.BindEnv(DBSslModeKey)
	viper.BindPFlag(DBSslModeKey, fs.Lookup(DBSslModeKey))
	viper.SetDefault(DBSslModeKey, "disable")
	viper.BindEnv(DBUrlKey)
	viper.BindPFlag(DBUrlKey, fs.Lookup(DBUrlKey))

	// graphql key & env
	viper.BindEnv(GraphqlEndpointKey)
	viper.BindPFlag(GraphqlEndpointKey, fs.Lookup(GraphqlEndpointKey))

	// fund key & env
	viper.BindEnv(FundWalletKey)
	viper.BindPFlag(FundWalletKey, fs.Lookup(FundWalletKey))
	viper.BindEnv(FundPoolAddressKey)
	viper.BindPFlag(FundPoolAddressKey, fs.Lookup(FundPoolAddressKey))
	viper.BindEnv(FundChainCodeKey)
	viper.BindPFlag(FundChainCodeKey, fs.Lookup(FundChainCodeKey))
	viper.BindEnv(FundRPCURLKey)
	viper.BindPFlag(FundRPCURLKey, fs.Lookup(FundRPCURLKey))
	viper.BindEnv(FundReconnectIntervalKey)
	viper.BindPFlag(FundReconnectIntervalKey, fs.Lookup(FundReconnectIntervalKey))
	viper.SetDefault(FundReconnectIntervalKey, 30)
	viper.BindEnv(FundBlockConfirmationKey)
	viper.BindPFlag(FundBlockConfirmationKey, fs.Lookup(FundBlockConfirmationKey))
	viper.SetDefault(FundBlockConfirmationKey, 12)
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
