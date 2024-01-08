package evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethereumClient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/kwilteam/kwil-db/core/types/chain"

	"github.com/ethereum/go-ethereum/core/types"
	ec "github.com/ethereum/go-ethereum/crypto"
)

type EthClient struct {
	ethclient *ethereumClient.Client
	chainCode chain.ChainCode
	endpoint  string
}

func New(ctx context.Context, endpoint string, chainCode chain.ChainCode) (*EthClient, error) {
	client, err := ethereumClient.DialContext(ctx, endpoint)
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

func (c *EthClient) ChainCode() chain.ChainCode {
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

func (c *EthClient) ChainID(ctx context.Context) (*big.Int, error) {
	return c.ethclient.ChainID(ctx)
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

	// suggest gas
	gasPrice, err := c.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	// set values
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func (c *EthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*chain.Header, error) {
	header, err := c.ethclient.HeaderByNumber(ctx, number)
	if err != nil {
		return nil, err
	}

	return &chain.Header{
		Hash:   header.Hash().Hex(),
		Height: header.Number.Int64(),
	}, nil
}

func (c *EthClient) SubscribeNewHead(ctx context.Context, ch chan<- chain.Header) (chain.Subscription, error) {

	ethHeaderChan := make(chan *types.Header)

	sub, err := c.ethclient.SubscribeNewHead(ctx, ethHeaderChan)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new EVM block headers: %v", err)
	}

	newSub := newEthSubscription(sub)

	// we will listen and convert the headers to our own headers/
	// this is simply a passthrough
	go func(ctx context.Context, ethHeaderChan <-chan *types.Header, ch chan<- chain.Header, sub *ethSubscription) {
		for {
			select {
			case ethHeader := <-ethHeaderChan:
				ch <- chain.Header{
					Height: ethHeader.Number.Int64(),
					Hash:   ethHeader.Hash().Hex(),
				}
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				newSub.errs <- err
			}
		}
	}(ctx, ethHeaderChan, ch, newSub)

	return newSub, nil
}
