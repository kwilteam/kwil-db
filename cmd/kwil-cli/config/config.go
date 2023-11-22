package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/prompt"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/internal/utils"

	"github.com/spf13/viper"
)

type KwilCliConfig struct {
	PrivateKey  *crypto.Secp256k1PrivateKey
	GrpcURL     string // TODO: change to maybe `RPCProvider` or `ProviderURL`
	ChainID     string
	TLSCertFile string // NOTE: since HTTP by default, this seems not use
}

func (c *KwilCliConfig) ToPersistedConfig() *kwilCliPersistedConfig {
	var privKeyHex string
	if c.PrivateKey != nil {
		privKeyHex = c.PrivateKey.Hex()
	}
	return &kwilCliPersistedConfig{
		PrivateKey:  privKeyHex,
		GrpcURL:     c.GrpcURL,
		ChainID:     c.ChainID,
		TLSCertFile: c.TLSCertFile,
	}
}

func (c *KwilCliConfig) Store() error {
	return PersistConfig(c)
}

func DefaultKwilCliPersistedConfig() *kwilCliPersistedConfig {
	return &kwilCliPersistedConfig{
		GrpcURL: "127.0.0.1:50051",
	}
}

// kwilCliPersistedConfig is the config that is used to persist the config file
// and also to work with viper(flags)
type kwilCliPersistedConfig struct {
	// NOTE: `mapstructure` is used by viper, name is same as the viper key name
	PrivateKey  string `mapstructure:"private_key" json:"private_key"`
	GrpcURL     string `mapstructure:"grpc_url" json:"grpc_url"`
	ChainID     string `mapstructure:"chain_id" json:"chain_id"`
	TLSCertFile string `mapstructure:"tls_cert_file" json:"tls_cert_file"`
}

func (c *kwilCliPersistedConfig) toKwilCliConfig() (*KwilCliConfig, error) {
	kwilConfig := &KwilCliConfig{
		GrpcURL:     c.GrpcURL,
		ChainID:     c.ChainID,
		TLSCertFile: c.TLSCertFile,
	}

	privateKey, err := crypto.Secp256k1PrivateKeyFromHex(c.PrivateKey)
	if err != nil {
		return kwilConfig, nil
	}

	kwilConfig.PrivateKey = privateKey

	return kwilConfig, nil
}

func PersistConfig(conf *KwilCliConfig) error {
	persistable := conf.ToPersistedConfig()

	jsonBytes, err := json.Marshal(persistable)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	file, err := utils.CreateOrOpenFile(defaultConfigFile)
	if err != nil {
		return fmt.Errorf("failed to create or open config file: %w", err)
	}

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate config file: %w", err)
	}

	_, err = file.Write(jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to write to config file: %w", err)
	}

	return nil
}

func LoadPersistedConfig() (*KwilCliConfig, error) {
	bts, err := utils.ReadOrCreateFile(defaultConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create or open config file: %w", err)
	}

	if len(bts) == 0 {
		fmt.Printf("config file is empty, creating new one")
		return &KwilCliConfig{}, nil
	}

	var conf kwilCliPersistedConfig
	err = json.Unmarshal(bts, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return conf.toKwilCliConfig()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// LoadCliConfig loads the config.
// The precedence order is following, each item takes precedence over the item below it:
//  1. flags
//  2. config file
//  3. default config
func LoadCliConfig() (*KwilCliConfig, error) {
	// NOTE: flags are already parsed, and viper also have the bind flags value
	// since flags has higher precedence, all below config values will be
	// overwritten by flags(if flags are set)

	// create default config and set to viper
	defaultCfg := DefaultKwilCliPersistedConfig()
	bs, err := json.Marshal(defaultCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	defaultConfig := bytes.NewReader(bs)
	viper.SetConfigType("json")
	if err := viper.MergeConfig(defaultConfig); err != nil {
		return nil, err
	}

	// NOTE: defaultConfigFile is set in init() in flags.go
	// read default config file if it exists
	// and override viper values from config file
	if fileExists(defaultConfigFile) {
		viper.SetConfigFile(defaultConfigFile)
		if err := viper.MergeInConfig(); err != nil {
			fmt.Printf("Error reading config file: %s\n", err)
			askAndDeleteConfig()
		}
	}

	// populate a new config with viper values(by viper key name)
	cfg := &kwilCliPersistedConfig{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg.toKwilCliConfig()
}

func askAndDeleteConfig() {
	askDelete := &prompt.Prompter{
		Label: fmt.Sprintf("Would you like to delete the corrupted config file at %s? (y/n) ", viper.ConfigFileUsed()),
	}

	response, err := askDelete.Run()
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return
	}

	if response != "y" {
		fmt.Println("Not deleting config file.  Using default values and/or flags.")
		return
	}

	err = os.Remove(viper.ConfigFileUsed())
	if err != nil {
		fmt.Printf("Error deleting config file: %s\n", err)
		return
	}
}
