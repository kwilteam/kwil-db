package evm

import (
	escrowCtr "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/escrow"
	tokenCtr "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/token"
	evmClient "github.com/kwilteam/kwil-db/core/chain/evm"
	"github.com/kwilteam/kwil-db/core/types/chain"
)

// TODO: Should this hold interfaces or explicit types???
type ethBridgeClient struct {
	*evmClient.EthClient
	*tokenCtr.Token
	*escrowCtr.Escrow
}

func New(endpoint string, chainCode chain.ChainCode, escrowAddress string) (*ethBridgeClient, error) {
	client, err := evmClient.New(endpoint, chainCode)
	if err != nil {
		return nil, err
	}
	escrow, err := escrowCtr.New(client, escrowAddress, chainCode.ToChainId())
	if err != nil {
		return nil, err
	}

	tokenAddress := escrow.TokenAddress()

	// if tokenAddress == "" {
	// 	tokenAddress = escrow.TokenAddress() // Does this work?
	// }

	token, err := tokenCtr.New(client, tokenAddress, chainCode.ToChainId())
	if err != nil {
		return nil, err
	}

	return &ethBridgeClient{
		EthClient: client,
		Token:     token,
		Escrow:    escrow,
	}, nil
}
