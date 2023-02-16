package config

import (
	"crypto/ecdsa"
	"fmt"
	"kwil/pkg/crypto"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/viper"
)

type CliConfig struct {
	Node struct {
		KwilProviderRpcUrl string `yaml:"rpc_url"`
	} `yaml:"node"`
	Wallet struct {
		PrivateKey string `yaml:"private_key"`
	} `yaml:"wallet"`
	ClientChain struct {
		Provider string `yaml:"provider"`
	}
}

// loadConfig loads the configuration from the config file.
// If the config file is not found, it will create a new one.
func LoadConfig() {
	viper.SetConfigName(DefaultConfigName)
	viper.SetConfigType(DefaultConfigType)
	viper.AddConfigPath(getDefaultConfigDir())

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			p := getDefaultConfigDir() + "/" + defaultConfigFileName()
			fmt.Println(p)

			_, err = createFileIfNeeded(p)
			if err != nil {
				fmt.Println("failed to create config dir: ")
				panic(err)
			}
		} else {
			fmt.Println("failed to read config file: ")
			panic(err)
		}

		err = viper.ReadInConfig()
		if err != nil {
			fmt.Println("failed to read config file after creating: ")
			panic(err)
		}
	}

	fillConfigStruct()
}

// TODO: I know we should use unmarshalling but the yaml package is giving me the most fucking unhelpful error
func fillConfigStruct() {
	Config.Node.KwilProviderRpcUrl = viper.GetString(KwilProviderRpcUrlKey)
	Config.Wallet.PrivateKey = viper.GetString(WalletPrivateKeyKey)
	Config.ClientChain.Provider = viper.GetString(ClientChainProviderRpcUrlKey)
}

func getUserRootDir() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return usr.HomeDir
}

func getDefaultConfigDir() string {
	return getUserRootDir() + "/" + DefaultConfigDir
}

func defaultConfigFileName() string {
	return DefaultConfigName + "." + DefaultConfigType
}

func createFileIfNeeded(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println(1)
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(2)
		return nil, err
	}

	return file, nil
}

func GetEcdsaPrivateKey() (*ecdsa.PrivateKey, error) {
	if Config.Wallet.PrivateKey == "" {
		return nil, fmt.Errorf("wallet private key is not set")
	}

	return crypto.ECDSAFromHex(Config.Wallet.PrivateKey)
}

func GetWalletAddress() (string, error) {
	ecdsaKey, err := GetEcdsaPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to get ecdsa key: %w", err)
	}

	return crypto.AddressFromPrivateKey(ecdsaKey)
}
