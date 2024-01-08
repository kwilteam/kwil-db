package evm

import (
	"context"

	evmClient "github.com/kwilteam/kwil-db/core/chain/evm"
	"github.com/kwilteam/kwil-db/core/types/chain"
	escrowCtr "github.com/kwilteam/kwil-db/oracles/tokenbridge/contracts/evm/escrow"
	tokenCtr "github.com/kwilteam/kwil-db/oracles/tokenbridge/contracts/evm/token"
)

// TODO: Should this hold interfaces or explicit types???
type ethBridgeClient struct {
	*evmClient.EthClient
	*tokenCtr.Token
	*escrowCtr.Escrow
}

func New(ctx context.Context, endpoint string, chainCode chain.ChainCode, escrowAddress string) (*ethBridgeClient, error) {
	client, err := evmClient.New(ctx, endpoint, chainCode)
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
