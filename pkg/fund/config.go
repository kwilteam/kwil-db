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
	ChainCode         int64             `mapstructure:"chain_code"`
	Wallet            *ecdsa.PrivateKey `mapstructure:"wallet"`
	TokenAddress      string            `mapstructure:"token_address"`
	PoolAddress       string            `mapstructure:"pool_address"`
	ValidatorAddress  string            `mapstructure:"validator_address"`
	Provider          string            `mapstructure:"provider"`
	ReConnectInterval int64             `mapstructure:"reconnect_interval"`
	BlockConfirmation int64             `mapstructure:"block_confirmation"`
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
		ChainCode:         chainCode,
		Wallet:            privateKey,
		PoolAddress:       fundingPool,
		ValidatorAddress:  validatorAddress,
		Provider:          ethProvider,
		TokenAddress:      tokenAddress,
		ReConnectInterval: reconnectionInterval,
		BlockConfirmation: requiredConfirmations,
	}, nil
}

func (c *Config) GetAccountAddress() string {
	return ec.PubkeyToAddress(c.Wallet.PublicKey).Hex()
}
