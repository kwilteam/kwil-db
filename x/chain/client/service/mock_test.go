package service_test

// this package is to contain the mock implementations
import (
	"context"
	"kwil/x/chain/client"
	"kwil/x/chain/client/dto"
	"kwil/x/chain/client/service"
	"kwil/x/chain/provider"
	providerDto "kwil/x/chain/provider/dto"
	"kwil/x/chain/types"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

type MockChainProvider struct {
	chainCode types.ChainCode
}

func (m *MockChainProvider) HeaderByNumber(ctx context.Context, number *big.Int) (*providerDto.Header, error) {
	var num int64
	if number == nil {
		num = STARTING_BLOCK
	} else {
		num = number.Int64()
	}

	return &providerDto.Header{
		Height: num,
		Hash:   "hash",
	}, nil
}

func (m *MockChainProvider) ChainCode() types.ChainCode {
	return m.chainCode
}

func (m *MockChainProvider) Endpoint() string {
	return "endpoint"
}

func newMockChainProvider() provider.ChainProvider {
	return &MockChainProvider{
		chainCode: CHAIN_CODE,
	}
}

func newMockChainClient() (client.ChainClient, error) {
	prov := newMockChainProvider()
	return service.NewChainClientExplicit(&dto.Config{
		Endpoint:              prov.Endpoint(),
		ChainCode:             2,
		ReconnectionInterval:  10,
		RequiredConfirmations: REQUIRED_CONFIRMATIONS,
	})

	/*
		return &chainClient{
			provider:              newMockChainProvider(),
			maxBlockInterval:      100,
			requiredConfirmations: 10,
			chainCode:             2,
		}
	*/
}

type mockSubscription struct {
	subbed bool
	errs   chan error
}

func (m *mockSubscription) Unsubscribe() {
	m.subbed = false
}

func (m *mockSubscription) Err() <-chan error {
	return m.errs
}

func newMockSubscription() *mockSubscription {
	return &mockSubscription{
		subbed: true,
		errs:   make(chan error),
	}
}

func (m *MockChainProvider) AsEthClient() (*ethclient.Client, error) {
	return nil, nil
}
