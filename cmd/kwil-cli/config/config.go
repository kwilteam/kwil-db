package config

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	common "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/prompt"
	"github.com/kwilteam/kwil-db/pkg/crypto"

	"github.com/spf13/viper"
)

type KwilCliConfig struct {
	PrivateKey        *ecdsa.PrivateKey
	GrpcURL           string
	ClientChainRPCURL string
}

func (c *KwilCliConfig) ToPeristedConfig() *kwilCliPersistedConfig {
	return &kwilCliPersistedConfig{
		PrivateKey:        crypto.HexFromECDSAPrivateKey(c.PrivateKey),
		GrpcURL:           c.GrpcURL,
		ClientChainRPCURL: c.ClientChainRPCURL,
	}
}

func (c *KwilCliConfig) Store() error {
	return PersistConfig(c)
}

type kwilCliPersistedConfig struct {
	PrivateKey        string `json:"private_key"`
	GrpcURL           string `json:"grpc_url"`
	ClientChainRPCURL string `json:"client_chain_rpc_url"`
}

func (c *kwilCliPersistedConfig) toKwilCliConfig() (*KwilCliConfig, error) {
	kwilConfig := &KwilCliConfig{
		GrpcURL:           c.GrpcURL,
		ClientChainRPCURL: c.ClientChainRPCURL,
	}

	privateKey, err := crypto.ECDSAFromHex(c.PrivateKey)
	if err != nil {
		return kwilConfig, nil
	}

	kwilConfig.PrivateKey = privateKey

	return kwilConfig, nil
}

func PersistConfig(conf *KwilCliConfig) error {
	persistable := conf.ToPeristedConfig()

	jsonBytes, err := json.Marshal(persistable)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	file, err := createOrOpenFile(DefaultConfigFile)
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

func createDirIfNeeded(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, os.ModePerm)
}

func readOrCreateFile(path string) ([]byte, error) {
	if err := createDirIfNeeded(path); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func createOrOpenFile(path string) (*os.File, error) {
	if err := createDirIfNeeded(path); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func LoadPersistedConfig() (*KwilCliConfig, error) {
	bts, err := readOrCreateFile(DefaultConfigFile)
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
		PrivateKey:        viper.GetString("private_key"),
		GrpcURL:           viper.GetString("grpc_url"),
		ClientChainRPCURL: viper.GetString("client_chain_rpc_url"),
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
