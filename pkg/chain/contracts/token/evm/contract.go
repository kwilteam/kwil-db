package evm

import (
	"fmt"
	"kwil/pkg/chain/contracts/token/evm/abi"
	"kwil/pkg/chain/provider"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type contract struct {
	client      provider.ChainProvider
	ctr         *abi.Erc20
	tokenName   string
	tokenSymbol string
	address     string
	decimals    uint8
	totalSupply *big.Int
	chainId     *big.Int
}

func New(provider provider.ChainProvider, chainId *big.Int, contractAddress string) (*contract, error) {
	client, err := provider.AsEthClient()
	if err != nil {
		return nil, err
	}

	ctr, err := abi.NewErc20(common.HexToAddress(contractAddress), client)
	if err != nil {
		return nil, err
	}

	tokenName, err := ctr.Name(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get token name: %d", err)
	}

	tokenSymbol, err := ctr.Symbol(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get token symbol: %d", err)
	}

	decimals, err := ctr.Decimals(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get token decimals: %d", err)
	}

	totalSupply, err := ctr.TotalSupply(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get token total supply: %d", err)
	}

	return &contract{
		client:      provider,
		ctr:         ctr,
		tokenName:   tokenName,
		tokenSymbol: tokenSymbol,
		decimals:    decimals,
		totalSupply: totalSupply,
		address:     contractAddress,
		chainId:     chainId,
	}, nil
}
