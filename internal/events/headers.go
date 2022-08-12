package events

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/kwilteam/kwil-db/internal/config"
	"github.com/rs/zerolog/log"
	"math/big"
	"time"
)

func (e *EventFeed) listenForBlockHeaders(ctx context.Context) (chan *big.Int, error) {

	headers := make(chan *types.Header)

	sub, err := e.EthClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, err
	}

	// This will be where the headers can be collected and processed in a buffer channel
	headerQueue := make(chan *big.Int, config.Conf.ClientChain.MaxBufferSize) // This channel will be used to pass headers
	notifierChannel := make(chan bool, 1)                                     // This channel is used to notify when headers are passed
	retChan := e.processBlockHeader(headerQueue, notifierChannel)             // This gets returned

	// Firing a goroutine to listen to incoming block headers
	go func() {
		lastBlock := big.NewInt(0)
		// Duration needs time in nanoseconds
		timeoutTime := time.Duration(1000000000 * config.Conf.ClientChain.BlockTimeout)
		for {
			select {
			case err := <-sub.Err():
				log.Fatal().Err(err).Msg("error on returned block header")

			case <-time.After(timeoutTime):
				// Websocket might have closed, re-launching
				e.log.Warn().Dur("timeout", timeoutTime).Msgf("web socket timeout, reconnecting")
				sub.Unsubscribe()
				sub = e.resubscribeEthClient(ctx, headers)

			case header := <-headers:

				// Checking here to make sure we don't miss any headers
				// First, checking if lastBlock is 0

				if len(lastBlock.Bits()) == 0 { // This is the first header received
					lastBlock = header.Number
					headerQueue <- header.Number
					notifierChannel <- true
				} else {
					// Set what we are expecting to receive
					expected := lastBlock.Add(lastBlock, big.NewInt(1))

					// Check to see if what we are receiving is what we are expecting
					if expected.Cmp(header.Number) == 0 {
						lastBlock = header.Number
						headerQueue <- header.Number
						notifierChannel <- true
					} else {
						// We need to recover the correct block here.
						// First, we need to find the difference between received and expected
						dif := big.NewInt(0)
						dif.Sub(expected, header.Number)

						// Check the difference between expected and received
						if expected.Cmp(header.Number) < 0 { // negative
							// If negative, then there are blocks that we missed

							// We can recover missed blocks here
							e.log.Warn().Str("expected", expected.String()).Str("received", lastBlock.String()).Msg("received block out of order")
							for i := new(big.Int).Set(expected); i.Cmp(header.Number) < 0; i.Add(i, big.NewInt(1)) {
								fmt.Println("Recovering block:", i)
								lastBlock = header.Number
								headerQueue <- i
								notifierChannel <- true
							}

						} else { // positive
							// If positive, then the client node sent the same block header twice (because for some reason it does that)
							// Do nothing in this instance, since we already received it
							e.log.Warn().Str("expected", expected.String()).Str("received", lastBlock.String()).Msg("received block twice")
							// Decrement lastBlock by dif
							lastBlock.Sub(lastBlock, dif)
						}
					}
				}
			}
		}
	}()
	return retChan, nil
}

// Make a function that will resubscribe to the block headers
func (e *EventFeed) resubscribeEthClient(ctx context.Context, headers chan *types.Header) ethereum.Subscription {
	sub, err := e.EthClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe to new block headers")
	}
	return sub
}

// This receives block headers and waits for them to reach a sufficient block height before pulling the data
func (e *EventFeed) processBlockHeader(ch chan *big.Int, nch chan bool) chan *big.Int {

	retChan := make(chan *big.Int, config.Conf.ClientChain.MaxBufferSize) // TODO: This can be an unbuffered channel, but I have made it buffered for testing purposes
	// Fire goroutine to listen
	go func() {
		for {
			// We will have a buffer channel that can store up to 50 block heights
			if len(ch) >= config.Conf.ClientChain.MinBlockHeight {
				// Once it is at 12, we can assume that this Ethereum fork is correct and pull the event data
				height := <-ch
				retChan <- height
				<-nch
				//log.Info().Str("block_height", height.String()).Msg("processing client block")
			} else {
				// Pull the notifier from the channel and restart the loop.
				// This is here so that it will hang until the header queue is long enough
				<-nch
			}
		}
	}()
	return retChan
}
