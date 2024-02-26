package evm

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/chain/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethereumclient "github.com/ethereum/go-ethereum/ethclient"
)

// client fulfills the dto.ChainProvider interface
type ethClient struct {
	ethclient *ethereumclient.Client
	chainCode types.ChainCode
	endpoint  string
}

func New(endpoint string, chainCode types.ChainCode) (*ethClient, error) {
	client, err := ethereumclient.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum node: %v", err)
	}

	return &ethClient{
		ethclient: client,
		chainCode: chainCode,
		endpoint:  endpoint,
	}, nil
}

func (c *ethClient) ChainCode() types.ChainCode {
	return c.chainCode
}

func (c *ethClient) AsEthClient() (*ethereumclient.Client, error) {
	if c.chainCode.ToChainId().Int64() == 0 {
		return nil, fmt.Errorf("unable to convert provider to ethclient: invalid chain code: %s", fmt.Sprint(c.chainCode))
	}
	return c.ethclient, nil
}

func (c *ethClient) Endpoint() string {
	return c.endpoint
}

func (c *ethClient) Close() error {
	c.ethclient.Close()
	return nil
}

func (c *ethClient) GetAccountNonce(ctx context.Context, address string) (uint64, error) {
	return c.ethclient.PendingNonceAt(ctx, common.HexToAddress(address))
}

func (c *ethClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return c.ethclient.SuggestGasPrice(ctx)
}
