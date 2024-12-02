package config

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers/prompt"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"

	"github.com/spf13/viper"
)

type KwilCliConfig struct {
	PrivateKey *crypto.Secp256k1PrivateKey
	Provider   string
	ChainID    string
}

// Identity returns the account ID, or nil if no private key is set. These are
// the bytes of the ethereum address.
func (c *KwilCliConfig) Identity() []byte {
	if c.PrivateKey == nil {
		return nil
	}
	signer := &auth.EthPersonalSigner{Key: *c.PrivateKey}
	return signer.Identity()
}

func (c *KwilCliConfig) ToPersistedConfig() *kwilCliPersistedConfig {
	var privKeyHex string
	if c.PrivateKey != nil {
		privKeyHex = hex.EncodeToString(c.PrivateKey.Bytes())
	}
	return &kwilCliPersistedConfig{
		PrivateKey: privKeyHex,
		Provider:   c.Provider,
		ChainID:    c.ChainID,
	}
}

func DefaultKwilCliPersistedConfig() *kwilCliPersistedConfig {
	return &kwilCliPersistedConfig{
		Provider: "http://127.0.0.1:8484",
	}
}

// kwilCliPersistedConfig is the config that is used to persist the config file
// and also to work with viper(flags)
type kwilCliPersistedConfig struct {
	// NOTE: `mapstructure` is used by viper, name is same as the viper key name
	PrivateKey string `mapstructure:"private_key" json:"private_key,omitempty"`
	Provider   string `mapstructure:"provider" json:"provider,omitempty"`
	ChainID    string `mapstructure:"chain_id" json:"chain_id,omitempty"`
}

func (c *kwilCliPersistedConfig) toKwilCliConfig() (*KwilCliConfig, error) {
	kwilConfig := &KwilCliConfig{
		Provider: c.Provider,
		ChainID:  c.ChainID,
	}

	// NOTE: so non private_key required cmds could be run
	if c.PrivateKey == "" {
		return kwilConfig, nil
	}

	// we should complain if the private key is configured and invalid
	privKeyBts, err := hex.DecodeString(c.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	privateKey, err := crypto.UnmarshalSecp256k1PrivateKey(privKeyBts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
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

	file, err := helpers.CreateOrOpenFile(configFile)
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
	bts, err := helpers.ReadOrCreateFile(configFile)
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

	// If the config file exists, override viper values from the config file
	// The config file is set through the flag and has the default value as
	// defaultConfigFile set in the init function in flags.go
	if fileExists(configFile) {
		viper.SetConfigFile(configFile)
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
