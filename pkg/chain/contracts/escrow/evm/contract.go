package evm

import (
	"github.com/kwilteam/kwil-db/pkg/chain/contracts/escrow/evm/abi"
	"github.com/kwilteam/kwil-db/pkg/chain/provider"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type contract struct {
	client  provider.ChainProvider
	ctr     *abi.Escrow
	token   string
	chainId *big.Int
	//providerAddress string
}

func New(provider provider.ChainProvider, contractAddress string) (*contract, error) {

	ethClient, err := provider.AsEthClient()
	if err != nil {
		return nil, err
	}

	ctr, err := abi.NewEscrow(common.HexToAddress(contractAddress), ethClient)
	if err != nil {
		return nil, err
	}

	tokAddr, err := ctr.EscrowToken(nil)
	if err != nil {
		return nil, err
	}

	return &contract{
		client:  provider,
		ctr:     ctr,
		token:   tokAddr.Hex(),
		chainId: provider.ChainCode().ToChainId(),
	}, nil
}

func (c *contract) TokenAddress() string {
	return c.token
}
