package escrow

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	evmClient "github.com/kwilteam/kwil-db/core/chain/evm"
	"github.com/kwilteam/kwil-db/oracles/tokenbridge/contracts/evm/escrow/abi"
)

type Escrow struct {
	client     *evmClient.EthClient
	ctr        *abi.Escrow
	tokenAddr  string
	escrowAddr string
	chainId    *big.Int
}

func New(client *evmClient.EthClient, escrowAddress string, chainId *big.Int) (*Escrow, error) {
	ctr, err := abi.NewEscrow(common.HexToAddress(escrowAddress), client.Backend())
	if err != nil {
		return nil, err
	}

	tokAddr, err := ctr.EscrowToken(&bind.CallOpts{Context: context.Background()})
	if err != nil {
		return nil, err
	}

	return &Escrow{
		client:     client,
		ctr:        ctr,
		tokenAddr:  tokAddr.Hex(),
		escrowAddr: escrowAddress,
		chainId:    chainId,
	}, nil
}

func (c *Escrow) TokenAddress() string {
	return c.tokenAddr
}
