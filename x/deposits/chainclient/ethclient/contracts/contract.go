package contracts

import (
	"context"

	"kwil/abi"
	ct "kwil/x/deposits/chainclient/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Contract interface {
	ReturnFunds(string, string, string) error
	GetDeposits(context.Context, int64, int64) ([]ct.Deposit, error)
}

type contract struct {
	ctr   *abi.Escrow
	token string
}

func New(client *ethclient.Client, addr string) (ct.Contract, error) {
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
	}, nil
}

func (c *contract) ReturnFunds(token, amount, recipient string) error {
	// TODO: implement
	return nil
}
