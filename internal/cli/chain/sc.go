package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/internal/cli/abi"
	"github.com/spf13/viper"
)

type Config struct {
	ChainId     string
	PrivateKey  string
	FundingPool string
	NodeAddress string
	EthProvider string
}

type Client struct {
	ChainID       string
	PrivateKey    *ecdsa.PrivateKey
	Address       *common.Address
	TokenAddr     *common.Address
	PoolAddr      *common.Address
	ValidatorAddr *common.Address
	abi           *abi.Abi
	erc20         *abi.Erc20
	client        *ethclient.Client
}

func NewClientV(v *viper.Viper) (*Client, error) {
	cfg := &Config{
		ChainId:     v.GetString("chain-id"),
		PrivateKey:  v.GetString("private-key"),
		FundingPool: v.GetString("funding-pool"),
		NodeAddress: v.GetString("node-address"),
		EthProvider: v.GetString("eth-provider"),
	}

	return NewClient(cfg)
}

func NewClient(c *Config) (*Client, error) {
	vAddr := common.HexToAddress(c.NodeAddress)
	poolAddr := common.HexToAddress(c.FundingPool)
	pkey, err := crypto.HexToECDSA(c.PrivateKey)
	if err != nil {
		fmt.Println("error parsing private key")
		return nil, err
	}

	addr := crypto.PubkeyToAddress(pkey.PublicKey)

	client, err := newEthClient(c.EthProvider)
	if err != nil {
		return nil, err
	}

	// load abi
	a, err := loadPoolSC(client, poolAddr)
	if err != nil {
		return nil, err
	}

	// get token address
	tAddr, err := a.EscrowToken(nil)
	if err != nil {
		return nil, err
	}

	// load erc20
	e, err := loadErc20SC(client, tAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		ChainID:       c.ChainId,
		PrivateKey:    pkey,
		Address:       &addr,
		abi:           a,
		erc20:         e,
		client:        client,
		TokenAddr:     &tAddr,
		PoolAddr:      &poolAddr,
		ValidatorAddr: &vAddr,
	}, nil
}

func (c *Client) GetBalance() (*big.Int, error) {
	return c.erc20.BalanceOf(nil, *c.Address)
}

func (c *Client) GetDepositBalance() (*big.Int, error) {
	return c.abi.Amounts(nil, *c.ValidatorAddr, *c.Address)
}

// approve approves the pool to spend the user's tokens
func (c *Client) Approve(amt *big.Int) error {

	auth, err := c.newEthAuth()
	if err != nil {
		return err
	}

	// approve the pool to spend the user's tokens
	tx, err := c.erc20.Approve(auth, *c.PoolAddr, amt)
	if err != nil {
		return err
	}

	fmt.Println("Success!")
	printTx(tx, c.ChainID)

	return nil
}

func (c *Client) GetAllowance() (*big.Int, error) {
	return c.erc20.Allowance(nil, *c.Address, *c.PoolAddr)
}

func (c *Client) Deposit(amt *big.Int, target string) error {
	auth, err := c.newEthAuth()
	if err != nil {
		return err
	}

	// convert target to common address
	tAddr := common.HexToAddress(target)

	// deposit the user's tokens
	tx, err := c.abi.Deposit(auth, tAddr, amt)
	if err != nil {
		return err
	}

	fmt.Println("Success!")
	printTx(tx, c.ChainID)

	return nil
}

func printTx(tx *types.Transaction, cid string) {
	switch cid {
	case "1":
		fmt.Printf("Etherscan: https://etherscan.io/tx/%s\n", tx.Hash().Hex())
	case "5":
		fmt.Printf("Etherscan: https://goerli.etherscan.io/tx/%s\n", tx.Hash().Hex())
	}
}

func (c *Client) newEthAuth() (*bind.TransactOpts, error) {
	// get pending nonce
	nonce, err := c.client.PendingNonceAt(context.Background(), *c.Address)
	if err != nil {
		return nil, err
	}

	// convert chain id to big int
	chainID, ok := new(big.Int).SetString(c.ChainID, 10)
	if !ok {
		return nil, fmt.Errorf("invalid chain id")
	}

	// create new auth
	auth, err := bind.NewKeyedTransactorWithChainID(c.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}

	// set values
	auth.Nonce = big.NewInt(int64(nonce))

	// suggest gas
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func loadPoolSC(client *ethclient.Client, pAddr common.Address) (*abi.Abi, error) {
	// load the pool contract
	return abi.NewAbi(pAddr, client)
}

func loadErc20SC(client *ethclient.Client, addr common.Address) (*abi.Erc20, error) {
	return abi.NewErc20(addr, client)
}

func newEthClient(provider string) (*ethclient.Client, error) {
	return ethclient.Dial("wss://" + provider)
}

func (c *Client) GetTokenName() (string, error) {
	return c.erc20.Name(nil)
}

func (c *Client) GetTokenSymbol() (string, error) {
	return c.erc20.Symbol(nil)
}

func (c *Client) GetValidatorAddress() string {
	return c.ValidatorAddr.Hex()
}

func (c *Client) GetPoolAddress() string {
	return c.PoolAddr.Hex()
}

func (c *Client) GetTokenDecimals() (uint8, error) {
	return c.erc20.Decimals(nil)
}

func (c *Client) Withdraw(amt *big.Int) error {
	auth, err := c.newEthAuth()
	if err != nil {
		return err
	}

	// withdraw the user's tokens
	tx, err := c.abi.RequestReturn(auth, *c.ValidatorAddr, amt)
	if err != nil {
		return err
	}

	fmt.Println("Success!")
	printTx(tx, c.ChainID)

	return nil
}

// divide the amount by 10^decimals
func (c *Client) ConvertToDecimal(amt *big.Int) (*big.Float, error) {
	// get decimals
	decimals, err := c.GetTokenDecimals()
	if err != nil {
		return nil, err
	}

	// convert to float
	f := new(big.Float).SetInt(amt)

	// divide by 10^decimals
	f.Quo(f, new(big.Float).SetInt(big.NewInt(int64(math.Pow10(int(decimals))))))

	return f, nil
}
