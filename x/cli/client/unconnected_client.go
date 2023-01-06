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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
)

type UnconnectedClient struct {
	ChainCode        chain.ChainCode
	PrivateKey       *ecdsa.PrivateKey
	Address          *common.Address
	TokenAddress     *common.Address
	PoolAddress      *common.Address
	ValidatorAddress *common.Address
	Escrow           escrow.EscrowContract
	Token            token.TokenContract
	ChainClient      chainClient.ChainClient
}

func NewUnconnectedClient(v *viper.Viper) (*UnconnectedClient, error) {
	chainCode := v.GetInt64("chain-code")
	fundingPool := common.HexToAddress(v.GetString("funding-pool"))
	nodeAddress := common.HexToAddress(v.GetString("node-address"))
	ethProvider := v.GetString("eth-provider")

	privateKey, err := crypto.HexToECDSA(v.GetString("private-key"))
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	chnClient, err := chainClientService.NewChainClientExplicit(&chainClientDto.Config{
		ChainCode:             chainCode,
		Endpoint:              ethProvider,
		ReconnectionInterval:  30,
		RequiredConfirmations: 12,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chain client: %v", err)
	}

	// escrow
	escrowCtr, err := escrow.New(chnClient, privateKey, fundingPool.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to create escrow contract: %v", err)
	}

	// erc20 address
	tokenAddress := common.HexToAddress(escrowCtr.TokenAddress())

	// erc20
	erc20Ctr, err := token.New(chnClient, privateKey, tokenAddress.Hex())
	if err != nil {
		return nil, fmt.Errorf("failed to create erc20 contract: %v", err)
	}

	return &UnconnectedClient{
		ChainCode:        chnClient.ChainCode(),
		PrivateKey:       privateKey,
		Address:          &address,
		PoolAddress:      &fundingPool,
		ValidatorAddress: &nodeAddress,
		ChainClient:      chnClient,
		Escrow:           escrowCtr,
		Token:            erc20Ctr,
		TokenAddress:     &tokenAddress,
	}, nil
}
