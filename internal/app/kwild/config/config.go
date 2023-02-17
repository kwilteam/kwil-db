package config

import (
	"crypto/ecdsa"
	"fmt"
	ec "github.com/ethereum/go-ethereum/crypto"
	"kwil/internal/pkg/config"
	"kwil/pkg/chain/client/dto"
	"kwil/pkg/log"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	EnvPrefix         = "KWILD"
	DefaultConfigDir  = ".kwild"
	DefaultConfigName = "config"
	DefaultConfigType = "yaml"
)

// viper keys
const (
	ServerListenAddrKey = "server.listen_addr"

	LogLevelKey       = "log.level"
	LogOutputPathsKey = "log.output_paths"

	DBHostKey     = "db.host"
	DBPortKey     = "db.port"
	DBUsernameKey = "db.username"
	DBPasswordKey = "db.password"
	DBDatabaseKey = "db.database"
	DBUrlKey      = "db.url"
	DBSslModeKey  = "db.sslmode"

	GraphqlAddr = "graphql.addr"

	FundWalletKey            = "fund.wallet"
	FundPoolAddressKey       = "fund.pool_address"
	FundChainCodeKey         = "fund.chain_code"
	FundRPCURLKey            = "fund.rpc_url"
	FundPublicRPCURLKey      = "fund.public_rpc_url"
	FundReconnectIntervalKey = "fund.reconnect_interval"
	FundBlockConfirmationKey = "fund.block_confirmation"

	GatewayAddrKey = "gateway.addr"
)

type FundConfig struct {
	Wallet      *ecdsa.PrivateKey `mapstructure:"wallet"`
	PoolAddress string            `mapstructure:"pool_address"`
	Chain       dto.Config        `mapstructure:",squash"`
}

func (c *FundConfig) GetAccountAddress() string {
	return ec.PubkeyToAddress(c.Wallet.PublicKey).Hex()
}

type GatewayConfig struct {
	Addr string `mapstructure:"addr"`
}

func (c *GatewayConfig) GetGraphqlUrl() string {
	graphqlUrl := c.Addr + "/graphql"
	return graphqlUrl
}

type GraphqlConfig struct {
	Addr string `mapstructure:"addr"`
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
	ListenAddr string `mapstructure:"listen_addr"`
}

type AppConfig struct {
	Server  ServerConfig   `mapstructure:"server"`
	Log     log.Config     `mapstructure:"log"`
	Fund    FundConfig     `mapstructure:"fund"`
	DB      PostgresConfig `mapstructure:"db"`
	Graphql GraphqlConfig  `mapstructure:"graphql"`
	Gateway GatewayConfig  `mapstructure:"gateway"`
}

var ConfigFile string

var defaultConfig = map[string]interface{}{
	"log": map[string]interface{}{
		"level":        "info",
		"output_paths": []string{"stdout"},
	},
	"db": map[string]interface{}{
		"port":    5432,
		"sslmode": "disable",
	},
	"server": map[string]interface{}{
		"listen_addr": "0.0.0.0:50051",
	},
	"graphql": map[string]interface{}{
		"addr": "localhost:8080",
	},
	"fund": map[string]interface{}{
		"reconnect_interval": 30,
		"block_confirmation": 12,
	},
	"gateway": map[string]interface{}{
		"addr": "localhost:8082",
	},
}

// BindGlobalFlags binds the global flags to the command.
func BindGlobalFlags(fs *pflag.FlagSet) {
	// server flags
	fs.String(ServerListenAddrKey, "", "the address of the Kwil server")

	// log flags
	fs.String(LogLevelKey, "", "the level of the Kwil log (default: info)")
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
	fs.String(GraphqlAddr, "", "the address of the Kwil graphql server")

	// fund flags
	fs.String(FundWalletKey, "", "your wallet private key")
	fs.String(FundPoolAddressKey, "", "the address of the funding pool")
	fs.String(FundChainCodeKey, "", "the chain code of the funding pool chain")
	fs.String(FundRPCURLKey, "", "the provider rpc url of the funding pool chain")
	fs.String(FundPublicRPCURLKey, "", "public provider rpc url of the funding pool chain for user onboarding")
	fs.Int64(FundReconnectIntervalKey, 0, "the reconnect interval of the funding pool")
	fs.Int64(FundBlockConfirmationKey, 0, "the block confirmation of the funding pool")

	// gateway flags
	fs.String(GatewayAddrKey, "", "the address of the Kwil gateway server")
}

// BindGlobalEnv binds the global flags to the environment variables.
func BindGlobalEnv(fs *pflag.FlagSet) {
	// node.endpoint maps to PREFIX_NODE_ENDPOINT
	viper.SetEnvPrefix(EnvPrefix)

	envs := []string{
		DBDatabaseKey,
		DBHostKey,
		DBPasswordKey,
		DBPortKey,
		DBSslModeKey,
		DBUrlKey,
		DBUsernameKey,
		FundBlockConfirmationKey,
		FundChainCodeKey,
		FundPoolAddressKey,
		FundReconnectIntervalKey,
		FundRPCURLKey,
		FundPublicRPCURLKey,
		FundWalletKey,
		GraphqlAddr,
		GatewayAddrKey,
		LogLevelKey,
		LogOutputPathsKey,
		ServerListenAddrKey,
	}

	for _, v := range envs {
		viper.BindEnv(v)
		viper.BindPFlag(v, fs.Lookup(v))
	}
}

func LoadConfig() (cfg *AppConfig, err error) {
	config.LoadConfig(defaultConfig, ConfigFile, EnvPrefix, DefaultConfigDir, DefaultConfigName, DefaultConfigType)

	if err = viper.Unmarshal(&cfg, viper.DecodeHook(config.StringPrivateKeyHookFunc())); err != nil {
		fmt.Fprintln(os.Stderr, "error unmarshal config file:", err)
	}
	return cfg, nil
}
