package evm

import (
	"crypto/ecdsa"
	"fmt"
	"kwil/x/contracts/token/evm/abi"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type contract struct {
	client      *ethclient.Client
	ctr         *abi.Erc20
	tokenName   string
	tokenSymbol string
	address     string
	decimals    uint8
	totalSupply *big.Int
	chainId     *big.Int
	privateKey  *ecdsa.PrivateKey
}

func New(client *ethclient.Client, chainId *big.Int, privateKey *ecdsa.PrivateKey, contractAddress string) (*contract, error) {

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
		client:      client,
		ctr:         ctr,
		tokenName:   tokenName,
		tokenSymbol: tokenSymbol,
		decimals:    decimals,
		totalSupply: totalSupply,
		address:     contractAddress,
		chainId:     chainId,
		privateKey:  privateKey,
	}, nil
}
