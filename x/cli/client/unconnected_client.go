package client

import (
	"crypto/ecdsa"
	"fmt"
	"kwil/x/chain"
	chainClient "kwil/x/chain/client"
	chainClientDto "kwil/x/chain/client/dto"
	chainClientService "kwil/x/chain/client/service"
	"kwil/x/contracts/escrow"
	"kwil/x/contracts/token"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
)

// an unconnected client is a client that is not made to connect to a kwil server, and only
// contains config and blockchain information

type UnconnectedClient interface {
	ChainCode() chain.ChainCode
	PrivateKey() *ecdsa.PrivateKey
	Address() string
	TokenAddress() string
	PoolAddress() string
	ValidatorAddress() string
	Escrow() escrow.EscrowContract
	Token() token.TokenContract
	ChainClient() chainClient.ChainClient
}

type unconnectedClient struct {
	chainCode        chain.ChainCode
	privateKey       *ecdsa.PrivateKey
	address          string
	tokenAddress     string
	poolAddress      string
	validatorAddress string
	escrow           escrow.EscrowContract
	token            token.TokenContract
	chainClient      chainClient.ChainClient
}

func NewUnconnectedClient(v *viper.Viper) (UnconnectedClient, error) {
	chainCode := v.GetInt64("chain-code")
	fundingPool := v.GetString("funding-pool")
	nodeAddress := v.GetString("node-address")
	ethProvider := v.GetString("eth-provider")

	privateKey, err := crypto.HexToECDSA(v.GetString("private-key"))
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	chnClient, err := chainClientService.NewChainClientExplicit(&chainClientDto.Config{
		ChainCode:             chainCode,
		Endpoint:              "wss://" + ethProvider,
		ReconnectionInterval:  30,
		RequiredConfirmations: 12,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chain client: %v", err)
	}

	// escrow
	escrowCtr, err := escrow.New(chnClient, privateKey, fundingPool)
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %v", err)
	}

	// erc20 address
	tokenAddress := escrowCtr.TokenAddress()

	// erc20
	erc20Ctr, err := token.New(chnClient, privateKey, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create erc20 contract: %v", err)
	}

	return &unconnectedClient{
		chainCode:        chnClient.ChainCode(),
		privateKey:       privateKey,
		address:          address,
		poolAddress:      fundingPool,
		validatorAddress: nodeAddress,
		chainClient:      chnClient,
		escrow:           escrowCtr,
		token:            erc20Ctr,
		tokenAddress:     tokenAddress,
	}, nil
}
