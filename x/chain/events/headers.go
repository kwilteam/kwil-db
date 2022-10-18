package events

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
)

func (ef *EventFeed) listenForBlockHeaders(ctx context.Context) (chan *big.Int, error) {
	headers := make(chan *types.Header)
	sub, err := ef.EthClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, err
	}

	lh, err := ef.ds.GetLastHeight()
	if err != nil {
		return nil, err
	}

	headerStore := HeaderQueue{
		queue: []*big.Int{},
		last:  lh,
	}

	retChan := make(chan *big.Int, ef.conf.GetBufferSize())

	// goroutine listens to new headers
	go func() {
		// Duration needs time in nanoseconds
		timeoutTime := time.Duration(1000000000 * ef.conf.GetBlockTimeout())
		for {
			select {
			case err := <-sub.Err():
				log.Warn().Err(err).Msg("unknown error from eth client, resubscribing")
				sub.Unsubscribe()
				sub = ef.resubscribeEthClient(ctx, headers)
			case <-time.After(timeoutTime):
				// Websocket might have closed, re-launching
				ef.log.Warn().Dur("timeout", timeoutTime).Msgf("web socket timeout, reconnecting")
				sub.Unsubscribe()
				sub = ef.resubscribeEthClient(ctx, headers)
			case header := <-headers:
				/*
					In this section, we will receive headers and store them in an array

					1.
					When a header is received, we will check to ensure it is one greater than last
					If it is, we will add it to the array
					If it is greater than one larger, we will add as many as necessary
					If it is less than one larger, we will do nothing

					2.
					If the HeaderQueue is <= than required confirmations, we do nothing
					If it is >, we will pull from the start of the queue and send it to the return channel
					We will pull until the queue is equal to the required confirmations

				*/

				ef.log.Debug().Msgf("received pre-commit header %s", header.Number.String())

				// 1.
				bi := big.NewInt(0)
				expected := bi.Add(headerStore.last, big.NewInt(1))
				if header.Number.Cmp(expected) == 0 {

					// This is the expected header
					ef.log.Debug().Msgf("received expected header %s", header.Number.String())
					headerStore.append(header.Number)

				} else if header.Number.Cmp(expected) > 0 {

					// received is greater than expected
					ef.log.Debug().Msgf("received header is greater than expected.  received: %s | expected: %s", header.Number.String(), expected.String())

					// Increasing the loop exit point since we need the received value as well
					endloop := header.Number.Add(header.Number, big.NewInt(1)) // looping through from expected to received
					for i := new(big.Int).Set(expected); i.Cmp(endloop) < 0; i.Add(i, big.NewInt(1)) {
						x := copyBigInt(i) //copy
						headerStore.append(x)
					}
				} else {
					ef.log.Debug().Msgf("received header is less than expected.  expected: %s | received: %s", header.Number.String(), expected.String())
					// do nothing here
				}
				// 2.
				for {
					// if queue is longer than required confirmations, we will pop and send
					if len(headerStore.queue) > ef.conf.GetReqConfirmations() {
						retChan <- headerStore.pop()
					} else {
						break
					}
				}

			}

		}

	}()

	return retChan, nil
}

// Make a function that will resubscribe to the block headers
func (ef *EventFeed) resubscribeEthClient(ctx context.Context, headers chan *types.Header) ethereum.Subscription {
	sub, err := ef.EthClient.SubscribeNewHead(ctx, headers)
	log.Debug().Msg("resubscribing to eth client")
	if err != nil {
		log.Warn().Err(err).Msg("failed to subscribe to new block headers, waiting 1 second and trying again")
		time.Sleep(1 * time.Second)
		sub, err = ef.EthClient.SubscribeNewHead(ctx, headers)
		if err != nil {
			log.Warn().Err(err).Msg("failed to subscribe to new block headers after 1 second, waiting 5 seconds and trying again")
			time.Sleep(5 * time.Second)
			sub, err = ef.EthClient.SubscribeNewHead(ctx, headers)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to subscribe to new block headers after 5 seconds, exiting")
			}
		}
	}
	return sub
}

type HeaderQueue struct {
	queue []*big.Int
	last  *big.Int
	mu    sync.Mutex
}

func (h *HeaderQueue) append(height *big.Int) {
	h.mu.Lock()
	if !h.isGreaterThanLast(height) {
		panic("height is not greater than last")
	}
	h.queue = append(h.queue, height)
	h.last = height
	h.mu.Unlock()
}

func (h *HeaderQueue) isGreaterThanLast(v *big.Int) bool {
	return v.Cmp(h.last) > 0
}

func (h *HeaderQueue) pop() *big.Int {
	h.mu.Lock()
	if len(h.queue) == 0 {
		return nil
	}
	ret := h.queue[0]
	h.queue = h.queue[1:]
	h.mu.Unlock()
	return ret
}

func copyBigInt(v *big.Int) *big.Int {
	return new(big.Int).Set(v)
}
