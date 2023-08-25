package config

import (
	"fmt"
	"path/filepath"

	cometCfg "github.com/cometbft/cometbft/config"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/spf13/viper"
)

type KwildConfig struct {
	RootDir    string
	PrivateKey *crypto.Ed25519PrivateKey

	AppCfg   *AppConfig       `mapstructure:"app"`
	ChainCfg *cometCfg.Config `mapstructure:"chain"`
	Logging  Logging          `mapstructure:"log"`
}

type Logging struct {
	LogLevel    string   `mapstructure:"log_level"`
	LogFormat   string   `mapstructure:"log_format"`
	OutputPaths []string `mapstructure:"output_paths"`
}

type AppConfig struct {
	GrpcListenAddress  string         `mapstructure:"grpc_laddr"`
	HttpListenAddress  string         `mapstructure:"http_laddr"`
	PrivateKey         string         `mapstructure:"private_key"`
	SqliteFilePath     string         `mapstructure:"sqlite_file_path"`
	ExtensionEndpoints []string       `mapstructure:"extension_endpoints"`
	WithoutGasCosts    bool           `mapstructure:"without_gas_costs"`
	WithoutNonces      bool           `mapstructure:"without_nonces"`
	SnapshotConfig     SnapshotConfig `mapstructure:"snapshots"`
	TLSCertFile        string         `mapstructure:"tls_cert_file"`
	TLSKeyFile         string         `mapstructure:"tls_key_file"`
	Hostname           string         `mapstructure:"hostname"`
	Log                log.Config
}

type SnapshotConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	RecurringHeight uint64 `mapstructure:"snapshot_heights"`
	MaxSnapshots    uint64 `mapstructure:"max_snapshots"`
	SnapshotDir     string `mapstructure:"snapshot_dir"`
}

func (cfg *KwildConfig) LoadKwildConfig(rootDir string) error {
	cfg.RootDir = rootDir
	cfg.ChainCfg.RootDir = rootDir
	cfg.ChainCfg.SetRoot(filepath.Join(rootDir, "abci"))

	cfgFile := filepath.Join(rootDir, "abci/config/config.toml")
	err := cfg.ParseConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	cfg.ConfigureLogging()
	cfg.ConfigureCerts()
	cfg.AppCfg.SqliteFilePath = rootify(cfg.AppCfg.SqliteFilePath, rootDir)
	cfg.AppCfg.SnapshotConfig.SnapshotDir = rootify(cfg.AppCfg.SnapshotConfig.SnapshotDir, rootDir)

	if err := cfg.ChainCfg.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid chain configuration data: %v", err)
	}

	return nil
}

func (cfg *KwildConfig) ParseConfig(cfgFile string) error {
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config: %v", err)
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("decoding config: %v", err)
	}
	return nil
}

func DefaultConfig() *KwildConfig {
	cfg := &KwildConfig{}
	cfg.ChainCfg = cometCfg.DefaultConfig()
	cfg.AppCfg = &AppConfig{
		GrpcListenAddress: "localhost:50051",
		HttpListenAddress: "localhost:8081",
		SqliteFilePath:    "data/kwil.db",
		WithoutGasCosts:   true,
		WithoutNonces:     false,
		SnapshotConfig: SnapshotConfig{
			Enabled:         false,
			RecurringHeight: uint64(10000),
			MaxSnapshots:    3,
			SnapshotDir:     "snapshots",
		},
	}
	return cfg
}

func (cfg *KwildConfig) ConfigureLogging() {
	// App Logging
	cfg.AppCfg.Log.Level = cfg.Logging.LogLevel
	cfg.AppCfg.Log.OutputPaths = cfg.Logging.OutputPaths

	// Chain Logging
	cfg.ChainCfg.LogLevel = cfg.Logging.LogLevel
	cfg.ChainCfg.LogFormat = cfg.Logging.LogFormat

}

func (cfg *KwildConfig) ConfigureCerts() {
	if cfg.AppCfg.TLSCertFile != "" {
		cfg.AppCfg.TLSCertFile = rootify(cfg.AppCfg.TLSCertFile, cfg.RootDir)
		cfg.ChainCfg.RPC.TLSCertFile = cfg.AppCfg.TLSCertFile
	}

	if cfg.AppCfg.TLSKeyFile != "" {
		cfg.AppCfg.TLSKeyFile = rootify(cfg.AppCfg.TLSKeyFile, cfg.RootDir)
		cfg.ChainCfg.RPC.TLSKeyFile = cfg.AppCfg.TLSKeyFile
	}
}

func rootify(path, rootDir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(rootDir, path)
}
