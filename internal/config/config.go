package config

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
)

var Conf types.Config

// Returns the current config
func GetConfig() *types.Config {
	return &Conf
}

// Function to load a file as the config
func LoadConfig(path string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName("dev")
	viper.SetConfigType("json")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err // Returning empty config if error occurs
	}

	err = viper.Unmarshal(&Conf)
	if err != nil {
		return err // Returning empty config if error occurs
	}

	err = loadABI(Conf.ClientChain.DepositContract.ABIPath)

	return err
}

// Will load an ABI from a file
func loadABI(path string) error {

	file, err := os.Open(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open ABI file")
	}
	abiJSON, err := abi.JSON(file)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse ABI file")
	}

	Conf.ClientChain.DepositContract.ABI = abiJSON
	return nil
}
