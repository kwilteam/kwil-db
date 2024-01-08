package token

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (c *Token) Allowance(ctx context.Context, owner, spender string) (*big.Int, error) {
	return c.ctr.Allowance(nil, common.HexToAddress(owner), common.HexToAddress(spender))
}

func (c *Token) Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (string, error) {
	auth, err := c.client.PrepareTxAuth(ctx, c.chainId, privateKey)
	if err != nil {
		return "", err
	}

	// create the transaction
	tx, err := c.ctr.Approve(auth, common.HexToAddress(spender), amount)
	if err != nil {
		return "", err
	}

	return tx.Hash().String(), nil
}

func (c *Token) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	return c.ctr.BalanceOf(&bind.CallOpts{
		Context: ctx,
	}, common.HexToAddress(address))
}
