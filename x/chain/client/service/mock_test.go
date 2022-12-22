package service_test

// this package is to contain the mock implementations
import (
	"context"
	"kwil/x/chain/client/dto"
	"kwil/x/chain/client/service"
	provider "kwil/x/chain/provider/dto"
	"math/big"
	"time"
)

type MockChainProvider struct {
}

func (m *MockChainProvider) HeaderByNumber(ctx context.Context, number *big.Int) (*provider.Header, error) {
	var num int64
	if number == nil {
		num = STARTING_BLOCK
	} else {
		num = number.Int64()
	}

	return &provider.Header{
		Height: num,
		Hash:   "hash",
	}, nil
}

func newMockChainProvider() provider.ChainProvider {
	return &MockChainProvider{}
}

func newMockChainClient() dto.ChainClient {

	interval := 10 * time.Second

	return service.NewChainClientNoConfig(newMockChainProvider(), 2, interval, REQUIRED_CONFIRMATIONS)

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
