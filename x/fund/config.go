package fund

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
	"kwil/x/chain/types"
)

// an unconnected client is a client that is not made to connect to a kwil server, and only
// contains config and blockchain information

type Config struct {
	ChainCode             int64
	PrivateKey            *ecdsa.PrivateKey
	TokenAddress          string
	PoolAddress           string
	ValidatorAddress      string
	Provider              string
	ReConnectionInterval  int64
	RequiredConfirmations int64
}

func NewConfig() (*Config, error) {
	// @yaiba TODO: make this less ugly, maybe from struct tags?
	chainCode := viper.GetInt64(types.ChainCodeFlag)
	fundingPool := viper.GetString(FundingPoolFlag)
	validatorAddress := viper.GetString(ValidatorAddressFlag)
	ethProvider := viper.GetString(types.EthProviderFlag)
	tokenAddress := viper.GetString(TokenAddressFlag)
	reconnectionInterval := viper.GetInt64(types.ReconnectionIntervalFlag)
	requiredConfirmations := viper.GetInt64(types.RequiredConfirmationsFlag)
	privateKey, err := crypto.HexToECDSA(viper.GetString(types.PrivateKeyFlag))
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}

	return &Config{
		ChainCode:             chainCode,
		PrivateKey:            privateKey,
		PoolAddress:           fundingPool,
		ValidatorAddress:      validatorAddress,
		Provider:              ethProvider,
		TokenAddress:          tokenAddress,
		ReConnectionInterval:  reconnectionInterval,
		RequiredConfirmations: requiredConfirmations,
	}, nil
}

func (c *Config) GetAccount() string {
	return crypto.PubkeyToAddress(c.PrivateKey.PublicKey).Hex()
}
