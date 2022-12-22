package service_test

import (
	"context"
	"fmt"
	provider "kwil/x/chain/provider/dto"
	"testing"
	"time"
)

var (
	EXPECTED_BLOCKS = []int64{101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 120, 112, 40, 121, 300, -1, -420, 4, 12, 2000}
)

const (
	REQUIRED_CONFIRMATIONS = 10
	STARTING_BLOCK         = 100
)

func (m *MockChainProvider) SubscribeNewHead(ctx context.Context, blocks chan<- provider.Header) (provider.Subscription, error) {

	sub := newMockSubscription()
	go func() {
		for _, number := range EXPECTED_BLOCKS {
			blocks <- provider.Header{
				Height: number,
				Hash:   "hash",
			}
		}

		sub.errs <- nil
	}()

	return sub, nil
}

// this tests the service's ability to receive blocks out of order and self-correct.  The blocks can be set in the EXPECTED_BLOCKS variable above
func Test_Service(t *testing.T) {
	// the test will start at height 100

	client := newMockChainClient()
	blocks := make(chan int64)
	err := client.Listen(context.Background(), blocks)
	if err != nil {
		t.Errorf("failed to listen to blocks: %v", err)
	}
	pos := 0
	shouldBreak := false
	currentBlock := EXPECTED_BLOCKS[0] - REQUIRED_CONFIRMATIONS
	for {
		select {
		case block := <-blocks:
			fmt.Println(pos)

			if block != currentBlock {
				t.Errorf("expected block %d, got %d", currentBlock, block)
				continue
			}
			t.Log("received expected block", block)
			currentBlock++
			pos++
		case <-time.After(1 * time.Second): // setting this so that this loop exits; the real consumer will never exit
			t.Logf("timed out")
			shouldBreak = true
		}
		if shouldBreak {
			break
		}
	}
}
