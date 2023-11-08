package client

import (
	"github.com/kwilteam/kwil-db/core/bridge/contracts"
	"github.com/kwilteam/kwil-db/core/chain"
)

type BridgeClient interface {
	ChainClient() chain.ChainClient
	TokenContract() contracts.TokenContract
	EscrowContract() contracts.EscrowContract
}
