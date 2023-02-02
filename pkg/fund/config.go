package fund

import (
	"crypto/ecdsa"
	"fmt"
	ec "github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
	"kwil/pkg/chain/types"
)

// an unconnected client is a client that is not made to connect to a kwil server, and only
// contains config and blockchain information

type Config struct {
	ChainCode            int64             `mapstructure:"chain_code"`
	PrivateKey           *ecdsa.PrivateKey `mapstructure:"private_key"`
	TokenAddress         string            `mapstructure:"token_address"`
	PoolAddress          string            `mapstructure:"pool_address"`
	ValidatorAddress     string            `mapstructure:"validator_address"`
	Provider             string            `mapstructure:"provider"`
	ReConnectionInterval int64             `mapstructure:"reconnection_interval"`
	RequiredConfirmation int64             `mapstructure:"required_confirmation"`
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
	privateKey, err := ec.HexToECDSA(viper.GetString(types.PrivateKeyFlag))
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}

	return &Config{
		ChainCode:            chainCode,
		PrivateKey:           privateKey,
		PoolAddress:          fundingPool,
		ValidatorAddress:     validatorAddress,
		Provider:             ethProvider,
		TokenAddress:         tokenAddress,
		ReConnectionInterval: reconnectionInterval,
		RequiredConfirmation: requiredConfirmations,
	}, nil
}

func (c *Config) GetAccountAddress() string {
	return ec.PubkeyToAddress(c.PrivateKey.PublicKey).Hex()
}
