// the token smart comtract is an abstraction for erc20 or equivalent tokens
package token

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	chainClient "kwil/pkg/chain/client"
	chainTypes "kwil/pkg/chain/types"
	"kwil/pkg/contracts/token/evm"
	types "kwil/pkg/contracts/token/types"
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
	Transfer(ctx context.Context, to string, amount *big.Int) (*types.TransferResponse, error)
	// approves the given amount of tokens to the given address
	Approve(ctx context.Context, spender string, amount *big.Int) (*types.ApproveResponse, error)
}

func New(chainClient chainClient.ChainClient, privateKey *ecdsa.PrivateKey, address string) (TokenContract, error) {
	switch chainClient.ChainCode() {
	case chainTypes.ETHEREUM, chainTypes.GOERLI:
		ethClient, err := chainClient.AsEthClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get ethclient from chain client: %d", err)
		}

		return evm.New(ethClient, chainClient.ChainCode().ToChainId(), privateKey, address)
	default:
		return nil, fmt.Errorf("unsupported chain code: %s", fmt.Sprint(chainClient.ChainCode()))
	}
}
