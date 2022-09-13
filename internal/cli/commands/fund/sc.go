package fund

// abi gen:
// abigen --abi=./abi/erc20.json --pkg=erc20 --out=erc20.go

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

type chainClient struct {
	cid           string
	pkey          *ecdsa.PrivateKey
	address       *common.Address
	abi           *abi.Abi
	erc20         *abi.Erc20
	client        *ethclient.Client
	tokenAddr     *common.Address
	poolAddr      *common.Address
	validatorAddr *common.Address
}

func newChainClient() (*chainClient, error) {
	// get client private key
	pk := viper.GetString("private-key")
	pa := viper.GetString("funding-pool")
	cid := viper.GetString("chain-id")
	va := viper.GetString("node-address")

	// convert va to common address
	vAddr := common.HexToAddress(va)

	// convert pa to common address
	poolAddr := common.HexToAddress(pa)

	// convert private key to public key
	pkey, err := crypto.HexToECDSA(pk)
	if err != nil {
		fmt.Println("error parsing private key")
		return nil, err
	}

	// get common address
	addr := crypto.PubkeyToAddress(pkey.PublicKey)

	// create eth client
	client, err := newEthClient()
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

	return &chainClient{
		cid:           cid,
		pkey:          pkey,
		address:       &addr,
		abi:           a,
		erc20:         e,
		client:        client,
		tokenAddr:     &tAddr,
		poolAddr:      &poolAddr,
		validatorAddr: &vAddr,
	}, nil
}

func (c *chainClient) getBalance() (*big.Int, error) {
	return c.erc20.BalanceOf(nil, *c.address)
}

func (c *chainClient) getDepositBalance() (*big.Int, error) {
	return c.abi.Amounts(nil, *c.validatorAddr, *c.address)
}

// approve approves the pool to spend the user's tokens
func (c *chainClient) approve(amt *big.Int) error {

	auth, err := c.newEthAuth()
	if err != nil {
		return err
	}

	// approve the pool to spend the user's tokens
	tx, err := c.erc20.Approve(auth, *c.poolAddr, amt)
	if err != nil {
		return err
	}

	fmt.Println("Success!")
	printTx(tx, c.cid)

	return nil
}

func (c *chainClient) getAllowance() (*big.Int, error) {
	return c.erc20.Allowance(nil, *c.address, *c.poolAddr)
}

func (c *chainClient) deposit(amt *big.Int, target string) error {
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
	printTx(tx, c.cid)

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

func (c *chainClient) newEthAuth() (*bind.TransactOpts, error) {
	// get pending nonce
	nonce, err := c.client.PendingNonceAt(context.Background(), *c.address)
	if err != nil {
		return nil, err
	}

	// convert chain id to big int
	chainID, ok := new(big.Int).SetString(c.cid, 10)
	if !ok {
		return nil, fmt.Errorf("invalid chain id")
	}

	// create new auth
	auth, err := bind.NewKeyedTransactorWithChainID(c.pkey, chainID)
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

func newEthClient() (*ethclient.Client, error) {
	// get eth provider
	ethProvider := viper.GetString("eth-provider")

	// connect to eth provider
	return ethclient.Dial("wss://" + ethProvider)
}

func (c *chainClient) getTokenName() (string, error) {
	return c.erc20.Name(nil)
}

func (c *chainClient) getTokenSymbol() (string, error) {
	return c.erc20.Symbol(nil)
}

func (c *chainClient) getValidatorAddress() string {
	return c.validatorAddr.Hex()
}

func (c *chainClient) getPoolAddress() string {
	return c.poolAddr.Hex()
}

func (c *chainClient) getTokenDecimals() (uint8, error) {
	return c.erc20.Decimals(nil)
}

func (c *chainClient) withdraw(amt *big.Int) error {
	auth, err := c.newEthAuth()
	if err != nil {
		return err
	}

	// withdraw the user's tokens
	tx, err := c.abi.RequestReturn(auth, *c.validatorAddr, amt)
	if err != nil {
		return err
	}

	fmt.Println("Success!")
	printTx(tx, c.cid)

	return nil
}

// divide the amount by 10^decimals
func (c *chainClient) convertToDecimal(amt *big.Int) (*big.Float, error) {
	// get decimals
	decimals, err := c.getTokenDecimals()
	if err != nil {
		return nil, err
	}

	// convert to float
	f := new(big.Float).SetInt(amt)

	// divide by 10^decimals
	f.Quo(f, new(big.Float).SetInt(big.NewInt(int64(math.Pow10(int(decimals))))))

	return f, nil
}
