package config

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/kwilteam/kwil-db/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

var Conf types.Config

// GetConfig Returns the current config
func GetConfig() *types.Config {
	return &Conf
}

// LoadConfig Function to load a file as the config
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

	err = initEnv(&Conf)
	if err != nil {
		return err
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

func initEnv(_ *types.Config) error {
	err := os.Setenv("TIMEOUT_TIME", strconv.Itoa(Conf.Api.TimeoutTime))
	if err != nil {
		return err
	}

	return nil
}
