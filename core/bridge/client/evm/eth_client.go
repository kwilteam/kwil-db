package evm

import (
	"github.com/kwilteam/kwil-db/core/bridge/contracts"
	escrowCtr "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/escrow"
	tokenCtr "github.com/kwilteam/kwil-db/core/bridge/contracts/evm/token"
	chainclt "github.com/kwilteam/kwil-db/core/chain"
	evmClient "github.com/kwilteam/kwil-db/core/chain/evm"
	"github.com/kwilteam/kwil-db/core/types/chain"
)

// TODO: Should this hold interfaces or explicit types???
type ethBridgeClient struct {
	client chainclt.ChainClient
	token  contracts.TokenContract
	escrow contracts.EscrowContract
}

func New(endpoint string, chainCode chain.ChainCode, tokenAddress string, escrowAddress string) (*ethBridgeClient, error) {
	client, err := evmClient.New(endpoint, chainCode)
	if err != nil {
		return nil, err
	}
	escrow, err := escrowCtr.New(client, escrowAddress, chainCode.ToChainId())
	if err != nil {
		return nil, err
	}

	if tokenAddress == "" {
		tokenAddress = escrow.TokenAddress()
	}

	token, err := tokenCtr.New(client, tokenAddress, chainCode.ToChainId())
	if err != nil {
		return nil, err
	}

	return &ethBridgeClient{
		client: client,
		token:  token,
		escrow: escrow,
	}, nil
}

// Provider Client methods

func (p *ethBridgeClient) ChainClient() chainclt.ChainClient {
	return p.client
}

func (p *ethBridgeClient) TokenContract() contracts.TokenContract {
	return p.token
}

func (p *ethBridgeClient) EscrowContract() contracts.EscrowContract {
	return p.escrow
}
