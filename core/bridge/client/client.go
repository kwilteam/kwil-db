package client

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/bridge/client/evm"
	"github.com/kwilteam/kwil-db/core/types/chain"
)

func New(endpoint string, chaincode chain.ChainCode, escrowAddress string) (TokenBridgeClient, error) {
	// func New(endpoint string, chaincode chain.ChainCode, escrowAddress string, tokenAddress string) (BridgeClient, error) {
	switch chaincode {
	case chain.ETHEREUM, chain.GOERLI:
		// return evm.New(endpoint, chaincode, tokenAddress, escrowAddress)
		return evm.New(endpoint, chaincode, escrowAddress)
	default:
		return nil, fmt.Errorf("unsupported chaincode: %s", chaincode)
	}
}
