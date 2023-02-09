package fund

import (
	"crypto/ecdsa"
	ec "github.com/ethereum/go-ethereum/crypto"
	"kwil/pkg/chain/client/dto"
)

// an unconnected client is a client that is not made to connect to a kwil server, and only
// contains config and blockchain information

type Config struct {
	Wallet      *ecdsa.PrivateKey `mapstructure:"wallet"`
	PoolAddress string            `mapstructure:"pool_address"`
	Chain       dto.Config        `mapstructure:",squash"`
}

func (c *Config) GetAccountAddress() string {
	return ec.PubkeyToAddress(c.Wallet.PublicKey).Hex()
}
