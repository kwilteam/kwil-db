package config

import (
	"encoding/json"
	"fmt"
	"os"

	common "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/prompt"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/utils"
	"github.com/spf13/viper"
)

type KwilCliConfig struct {
	PrivateKey crypto.PrivateKey
	GrpcURL    string
}

func (c *KwilCliConfig) ToPersistedConfig() *kwilCliPersistedConfig {
	var privKeyHex string
	if c.PrivateKey != nil {
		privKeyHex = c.PrivateKey.Hex()
	}
	return &kwilCliPersistedConfig{
		PrivateKey: privKeyHex,
		GrpcURL:    c.GrpcURL,
	}
}

func (c *KwilCliConfig) Store() error {
	return PersistConfig(c)
}

type kwilCliPersistedConfig struct {
	PrivateKey string `json:"private_key"`
	GrpcURL    string `json:"grpc_url"`
}

func (c *kwilCliPersistedConfig) toKwilCliConfig() (*KwilCliConfig, error) {
	kwilConfig := &KwilCliConfig{
		GrpcURL: c.GrpcURL,
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

	file, err := utils.CreateOrOpenFile(DefaultConfigFile)
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
	bts, err := utils.ReadOrCreateFile(DefaultConfigFile)
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

func LoadCliConfig() (*KwilCliConfig, error) {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Config file not found. Using default values and/or flags.  To create a config file, run 'kwil-cli configure'\n")
		} else {
			fmt.Printf("Error reading config file: %s\n", err)
			askAndDeleteConfig()
		}
	}

	innerConf := &kwilCliPersistedConfig{
		PrivateKey: viper.GetString("private_key"),
		GrpcURL:    viper.GetString("grpc_url"),
	}
	return innerConf.toKwilCliConfig()
}

func askAndDeleteConfig() {
	askDelete := &common.Prompter{
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
