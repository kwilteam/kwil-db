package escrow

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (c *Escrow) Deposit(ctx context.Context, amount *big.Int, privateKey *ecdsa.PrivateKey) (string, error) {

	auth, err := c.client.PrepareTxAuth(ctx, c.chainId, privateKey)
	if err != nil {
		return "", err
	}

	res, err := c.ctr.Deposit(auth, amount)
	if err != nil {
		return "", err
	}
	return res.Hash().String(), nil
}

func (c *Escrow) Balance(ctx context.Context, address string) (*big.Int, error) {
	return c.ctr.Balance(&bind.CallOpts{
		Context: ctx,
	}, common.HexToAddress(address))
}
