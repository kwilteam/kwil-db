package token

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kwilteam/kwil-db/core/bridge/contracts/evm/token/abi"
	evmClient "github.com/kwilteam/kwil-db/core/chain/evm"
)

type Token struct {
	client      *evmClient.EthClient
	ctr         *abi.Erc20
	tokenName   string
	tokenSymbol string
	address     string
	decimals    uint8
	totalSupply *big.Int
	chainId     *big.Int
}

func New(client *evmClient.EthClient, tokenAddress string, chainId *big.Int) (*Token, error) {
	ctr, err := abi.NewErc20(common.HexToAddress(tokenAddress), client.Backend())
	if err != nil {
		return nil, err
	}

	name, err := ctr.Name(nil)
	if err != nil {
		return nil, err
	}

	symbol, err := ctr.Symbol(nil)
	if err != nil {
		return nil, err
	}

	decimals, err := ctr.Decimals(nil)
	if err != nil {
		return nil, err
	}

	totalSupply, err := ctr.TotalSupply(nil)
	if err != nil {
		return nil, err
	}

	return &Token{
		client:      client,
		ctr:         ctr,
		tokenName:   name,
		tokenSymbol: symbol,
		address:     tokenAddress,
		decimals:    decimals,
		totalSupply: totalSupply,
		chainId:     chainId,
	}, nil
}
