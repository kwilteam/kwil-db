// the token smart comtract is an abstraction for erc20 or equivalent tokens
package token

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/pkg/chain/contracts/token/evm"
	types2 "kwil/pkg/chain/contracts/token/types"
	"kwil/pkg/chain/provider"
	chainTypes "kwil/pkg/chain/types"
	"kwil/pkg/log"
	"kwil/pkg/utils/retry"
	"math/big"
)

type TokenContract interface {
	// returns the name of the token
	Name() string
	// returns the symbol of the token
	Symbol() string
	// returns the total supply of the token
	TotalSupply() *big.Int
	// returns the balance of the given address
	BalanceOf(ctx context.Context, address string) (*big.Int, error)
	// returns the allowance of the given address
	Allowance(ctx context.Context, owner, spender string) (*big.Int, error)
	// returns the address of the token
	Address() string
	// returns the decimals of the token
	Decimals() uint8

	// transfers the given amount of tokens to the given address
	Transfer(ctx context.Context, to string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types2.TransferResponse, error)
	// approves the given amount of tokens to the given address
	Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types2.ApproveResponse, error)
}

func New(chainProvider provider.ChainProvider, address string, opts ...TokenOpts) (TokenContract, error) {
	var ctr TokenContract
	var err error

	switch chainProvider.ChainCode() {
	case chainTypes.ETHEREUM, chainTypes.GOERLI:
		ctr, err = evm.New(chainProvider, chainProvider.ChainCode().ToChainId(), address)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainProvider.ChainCode()))
	}

	if err != nil {
		return nil, err
	}

	return newRetry(ctr, opts...), nil
}

// internal token struct for building the retry mechanism
type token struct {
	ctr TokenContract
	log log.Logger
}

func newRetry(contract TokenContract, opts ...TokenOpts) TokenContract {
	t := &token{
		ctr: contract,
		log: log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (r *token) retry(ctx context.Context, fn func() error) error {
	return retry.Retry(func() error {
		return fn()
	},
		retry.WithContext(ctx),
		retry.WithLogger(r.log),
		retry.WithMin(1),
		retry.WithMax(10),
		retry.WithFactor(2),
	)
}

func (r *token) Name() string {
	return r.ctr.Name()
}

func (r *token) Symbol() string {
	return r.ctr.Symbol()
}

func (r *token) TotalSupply() *big.Int {
	return r.ctr.TotalSupply()
}

func (r *token) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	var balance *big.Int
	err := r.retry(ctx, func() error {
		var err error
		balance, err = r.ctr.BalanceOf(ctx, address)
		return err
	})
	return balance, err
}

func (r *token) Allowance(ctx context.Context, owner, spender string) (*big.Int, error) {
	var allowance *big.Int
	err := r.retry(ctx, func() error {
		var err error
		allowance, err = r.ctr.Allowance(ctx, owner, spender)
		return err
	})

	return allowance, err
}

func (r *token) Address() string {
	return r.ctr.Address()
}

func (r *token) Decimals() uint8 {
	return r.ctr.Decimals()
}

func (r *token) Transfer(ctx context.Context, to string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types2.TransferResponse, error) {
	var resp *types2.TransferResponse
	err := r.retry(ctx, func() error {
		var err error
		resp, err = r.ctr.Transfer(ctx, to, amount, privateKey)
		return err
	})

	return resp, err
}

func (r *token) Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types2.ApproveResponse, error) {
	var resp *types2.ApproveResponse
	err := r.retry(ctx, func() error {
		var err error
		resp, err = r.ctr.Approve(ctx, spender, amount, privateKey)
		return err
	})

	return resp, err
}
