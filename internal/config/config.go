package config

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/kwilteam/kwil-db/internal/utils/files"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strconv"
<<<<<<< HEAD
	"strings"
=======
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
)

func Init(c *Config) error {
	err := initEnv(c)
	if err != nil {
		return err
	}

<<<<<<< HEAD
	initFriends(c)

	return nil
}
=======
// LoadConfig Function to load a file as the config
func LoadConfig(path string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName("dev")
	viper.SetConfigType("json")
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5

// Will load an ABI from a file
func (c *Config) loadABI(path string) error {
	file, err := files.LoadFileFromRoot(path)
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

<<<<<<< HEAD
	return nil
}
=======
	err = Init(&Conf)

	loadABI(Conf.ClientChain.DepositContract.ABIPath)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5

func initFriends(c *Config) {
	// loop through the Friendlist and add each friend to the Friends map
	c.Friends = make(map[string]bool)
	for _, friend := range c.Friendlist {
		c.Friends[friend] = true
	}
}

<<<<<<< HEAD
// LoadConfig will load the config file
func LoadConfig(f string) (*Config, error) {
	// f is of the format "config.json"
	// we need to split the file name from the extension

	viper.SetConfigFile(f)

	var c Config
=======
func Init(c *types.Config) error {
	err := initEnv(c)
	if err != nil {
		return err
	}

	initFriends(c)

	return nil
}

// Will load an ABI from a file
func loadABI(path string) error {
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5

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

func initEnv(c *types.Config) error {
	err := os.Setenv("TIMEOUT_TIME", strconv.Itoa(c.Api.TimeoutTime))
	if err != nil {
		return err
	}

	return nil
}

func initFriends(c *types.Config) {
	// loop through the Friendlist and add each friend to the Friends map
	c.Friends = make(map[string]bool)
	for _, friend := range c.Friendlist {
		c.Friends[friend] = true
	}
}
