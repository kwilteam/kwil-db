package client

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethereumClient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/internal/chain/types"

	ec "github.com/ethereum/go-ethereum/crypto"
)

type EthClient struct {
	ethclient *ethereumClient.Client
	chainCode types.ChainCode
	endpoint  string
}

func New(endpoint string, chainCode types.ChainCode) (*EthClient, error) {
	client, err := ethereumClient.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	return &EthClient{
		ethclient: client,
		chainCode: chainCode,
		endpoint:  endpoint,
	}, nil
}

func (c *EthClient) Backend() *ethereumClient.Client {
	return c.ethclient
}

func (c *EthClient) ChainCode() types.ChainCode {
	return c.chainCode
}

func (c *EthClient) Endpoint() string {
	return c.endpoint
}

func (c *EthClient) Close() error {
	c.ethclient.Close()
	return nil
}

func (c *EthClient) GetAccountNonce(ctx context.Context, address string) (uint64, error) {
	return c.ethclient.PendingNonceAt(ctx, common.HexToAddress(address))
}

func (c *EthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return c.ethclient.SuggestGasPrice(ctx)
}

func (c *EthClient) PrepareTxAuth(ctx context.Context, chainId *big.Int, privateKey *ecdsa.PrivateKey) (*bind.TransactOpts, error) {
	addr := ec.PubkeyToAddress(privateKey.PublicKey)

	// get pending nonce
	nonce, err := c.GetAccountNonce(ctx, addr.Hex())
	if err != nil {
		return nil, err
	}

	// create new auth
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return nil, err
	}

	// set values
	auth.Nonce = big.NewInt(int64(nonce))

	// suggest gas
	gasPrice, err := c.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}
