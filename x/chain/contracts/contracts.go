package contracts

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"kwil/abi"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Account interface {
	GetPrivateKey() (*ecdsa.PrivateKey, error)
	GetAddress() *common.Address
}

type ContractClient struct {
	acc     Account
	client  *ethclient.Client
	chainID string
	escrow  *abi.Escrow
	erc20   *abi.Erc20
}

func NewContractClient(acc Account, client *ethclient.Client, escrowAddr, chainID string) (*ContractClient, error) {
	// get address from private key

	escrow, err := loadEscrowSC(client, common.HexToAddress(escrowAddr))
	if err != nil {
		return nil, err
	}

	erc20Addr, err := escrow.EscrowToken(nil)
	if err != nil {
		return nil, err
	}

	erc20, err := loadErc20SC(client, erc20Addr)
	if err != nil {
		return nil, err
	}

	return &ContractClient{
		acc:     acc,
		client:  client,
		chainID: chainID,
		escrow:  escrow,
		erc20:   erc20,
	}, nil
}

func (c *ContractClient) newAuth(ctx context.Context) (*bind.TransactOpts, error) {
	// get pending nonce
	nonce, err := c.client.PendingNonceAt(ctx, *c.acc.GetAddress())
	if err != nil {
		return nil, err
	}

	// convert chain id to big int
	chainID, ok := new(big.Int).SetString(c.chainID, 10)
	if !ok {
		return nil, fmt.Errorf("invalid chain id")
	}

	// retrieve private key
	pk, err := c.acc.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	// create new auth
	auth, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	if err != nil {
		return nil, err
	}

	// set values
	auth.Nonce = big.NewInt(int64(nonce))

	// suggest gas
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func (c *ContractClient) ReturnFunds(ctx context.Context, recipient common.Address, amount, fee *big.Int) (*types.Transaction, error) {
	auth, err := c.newAuth(ctx)
	if err != nil {
		return nil, err
	}

	return c.escrow.ReturnDeposit(auth, recipient, amount, fee)
}

func loadEscrowSC(client *ethclient.Client, pAddr common.Address) (*abi.Escrow, error) {
	// load the pool contract
	return abi.NewEscrow(pAddr, client)
}

func loadErc20SC(client *ethclient.Client, addr common.Address) (*abi.Erc20, error) {
	return abi.NewErc20(addr, client)
}
