package evm

import (
	"kwil/pkg/contracts/escrow/evm/abi"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type contract struct {
	client  *ethclient.Client
	ctr     *abi.Escrow
	token   string
	chainId *big.Int
	//providerAddress string
}

func New(client *ethclient.Client, chainId *big.Int, contractAddress string) (*contract, error) {

	ctr, err := abi.NewEscrow(common.HexToAddress(contractAddress), client)
	if err != nil {
		return nil, err
	}

	tokAddr, err := ctr.EscrowToken(nil)
	if err != nil {
		return nil, err
	}

	return &contract{
		client:  client,
		ctr:     ctr,
		token:   tokAddr.Hex(),
		chainId: chainId,
	}, nil
}

func (c *contract) TokenAddress() string {
	return c.token
}
