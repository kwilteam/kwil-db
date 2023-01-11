package client

import (
	"crypto/ecdsa"
	"fmt"
	chainClient "kwil/x/chain/client"
	"kwil/x/contracts/escrow"
	"kwil/x/contracts/token"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
)

// an unconnected client is a client that is not made to connect to a kwil server, and only
// contains config and blockchain information

type ClientConfig struct {
	ChainCode        int64
	PrivateKey       *ecdsa.PrivateKey
	Address          string
	TokenAddress     string
	PoolAddress      string
	ValidatorAddress string
	Escrow           escrow.EscrowContract
	Token            token.TokenContract
	ChainClient      chainClient.ChainClient
}

func NewClientConfig(v *viper.Viper) (*ClientConfig, error) {
	chainCode := v.GetInt64("chain-code")
	fundingPool := v.GetString("funding-pool")
	nodeAddress := v.GetString("node-address")

	privateKey, err := crypto.HexToECDSA(v.GetString("private-key"))
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	return &ClientConfig{
		ChainCode:        chainCode,
		PrivateKey:       privateKey,
		Address:          address,
		PoolAddress:      fundingPool,
		ValidatorAddress: nodeAddress,
	}, nil
}
