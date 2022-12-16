package contracts

import (
	"math/big"

	"kwil/abi"
	ct "kwil/x/deposits_old/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type contract struct {
	ctr   *abi.Escrow
	token string
	cid   *big.Int
}

func New(client *ethclient.Client, addr string, chainID *big.Int) (ct.Contract, error) {
	ctr, err := abi.NewEscrow(common.HexToAddress(addr), client)
	if err != nil {
		return nil, err
	}

	tokAddr, err := ctr.EscrowToken(nil)
	if err != nil {
		return nil, err
	}

	return &contract{
		ctr:   ctr,
		token: tokAddr.Hex(),
		cid:   chainID,
	}, nil
}
