package escrow

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	evmClient "github.com/kwilteam/kwil-db/internal/chain/provider/evm/client"
	"github.com/kwilteam/kwil-db/internal/chain/provider/evm/contracts/escrow/abi"
)

type EscrowContract struct {
	client    *evmClient.EthClient
	ctr       *abi.Escrow
	tokenAddr string
	chainId   *big.Int
}

func New(client *evmClient.EthClient, escrowAddress string, chainId *big.Int) (*EscrowContract, error) {
	ctr, err := abi.NewEscrow(common.HexToAddress(escrowAddress), client.Backend())
	if err != nil {
		return nil, err
	}

	tokAddr, err := ctr.EscrowToken(nil)
	if err != nil {
		return nil, err
	}

	return &EscrowContract{
		client:    client,
		ctr:       ctr,
		tokenAddr: tokAddr.Hex(),
		chainId:   chainId,
	}, nil
}

func (c *EscrowContract) TokenAddress() string {
	return c.tokenAddr
}
