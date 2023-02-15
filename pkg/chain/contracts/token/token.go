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
	BalanceOf(address string) (*big.Int, error)
	// returns the allowance of the given address
	Allowance(owner, spender string) (*big.Int, error)
	// returns the address of the token
	Address() string
	// returns the decimals of the token
	Decimals() uint8

	// transfers the given amount of tokens to the given address
	Transfer(ctx context.Context, to string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types2.TransferResponse, error)
	// approves the given amount of tokens to the given address
	Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (*types2.ApproveResponse, error)
}

func New(chainProvider provider.ChainProvider, address string) (TokenContract, error) {
	switch chainProvider.ChainCode() {
	case chainTypes.ETHEREUM, chainTypes.GOERLI:
		return evm.New(chainProvider, chainProvider.ChainCode().ToChainId(), address)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainProvider.ChainCode()))
	}
}
