package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/kwilteam/kwil-db/internal/chain/utils"
)

func Init(c *Config) error {
	err := initEnv(c)
	if err != nil {
		return err
	}

	initFriends(c)

	return nil
}

// Will load an ABI from a file
func (c *Config) loadABI(path string) error {
	file, err := utils.LoadFileFromRoot(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open ABI file")
	}

	// unmarshall file into ABI
	abiJSON, err := abi.JSON(strings.NewReader(string(file)))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshall ABI")
	}

	c.ClientChain.DepositContract.ABI = abiJSON
	return nil
}

func initEnv(c *Config) error {
	err := os.Setenv("TIMEOUT_TIME", strconv.Itoa(c.Api.TimeoutTime))
	if err != nil {
		return err
	}

	return nil
}

func initFriends(c *Config) {
	// loop through the Friendlist and add each friend to the Friends map
	c.Friends = make(map[string]bool)
	for _, friend := range c.Friendlist {
		c.Friends[friend] = true
	}
}

// LoadConfig will load the config file
func LoadConfig(f string) (*Config, error) {
	// f is of the format "config.json"
	// we need to split the file name from the extension

	viper.SetConfigFile(f)

	var c Config

	err := viper.ReadInConfig()
	if err != nil {
		return &c, err // Returning empty config if error occurs
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		return &c, err // Returning empty config if error occurs
	}

	err = Init(&c)

	c.loadABI(c.ClientChain.DepositContract.ABIPath)

	return &c, err
}
