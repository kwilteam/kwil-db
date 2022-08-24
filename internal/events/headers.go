package events

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/kwilteam/kwil-db/internal/config"
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

	retChan := make(chan *big.Int, config.Conf.ClientChain.MaxBufferSize)

	// goroutine listents to new headers
	go func() {
		// Duration needs time in nanoseconds
		timeoutTime := time.Duration(1000000000 * config.Conf.ClientChain.BlockTimeout)
		for {
			select {
			case err := <-sub.Err():
				log.Fatal().Err(err).Msg("failed to subscribe to new block headers")
			case <-time.After(timeoutTime):
				// Websocket might have closed, re-launching
				ef.log.Warn().Dur("timeout", timeoutTime).Msgf("web socket timeout, reconnecting")
				sub.Unsubscribe()
				sub = ef.resubscribeEthClient(ctx, headers)
			case header := <-headers:
				// IF SHIT IS BREAKING IT IS PROBABLY HERE

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
					headerStore.append(header.Number)

				} else if header.Number.Cmp(expected) > 0 {

					// received is greater than expected
					ef.log.Debug().Msgf("received header is greater than expected.  expected: %s | received: %s", header.Number.String(), expected.String())

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
					if len(headerStore.queue) > config.Conf.ClientChain.RequiredConfirmations {
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
func (e *EventFeed) resubscribeEthClient(ctx context.Context, headers chan *types.Header) ethereum.Subscription {
	sub, err := e.EthClient.SubscribeNewHead(ctx, headers)
	log.Debug().Msg("resubscribing to eth client")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe to new block headers")
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

/*func (ef *EventFeed) listenForBlockHeaders(ctx context.Context) (chan *big.Int, error) {

	headers := make(chan *types.Header)

	sub, err := ef.EthClient.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, err
	}

	// This will be where the headers can be collected and processed in a buffer channel
	headerQueue := make(chan *big.Int, config.Conf.ClientChain.MaxBufferSize) // This channel will be used to pass headers
	notifierChannel := make(chan bool, 1)                                     // This channel is used to notify when headers are passed
	retChan := ef.processBlockHeader(headerQueue, notifierChannel)            // This gets returned

	// Firing a goroutine to listen to incoming block headers
	go func() {
		lastBlock, err := ef.ds.GetLastHeight()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get last block height")
		}
		// Duration needs time in nanoseconds
		timeoutTime := time.Duration(1000000000 * config.Conf.ClientChain.BlockTimeout)
		for {
			select {
			case err := <-sub.Err():
				log.Fatal().Err(err).Msg("error on returned block header")

			case <-time.After(timeoutTime):
				// Websocket might have closed, re-launching
				ef.log.Warn().Dur("timeout", timeoutTime).Msgf("web socket timeout, reconnecting")
				sub.Unsubscribe()
				sub = ef.resubscribeEthClient(ctx, headers)

			case header := <-headers:
				/* placeholder
				########################################################################################################################################################
				########################################################################################################################################################
				########################################################################################################################################################
				########################################################################################################################################################


				// Checking here to make sure we don't miss any headers
				// First, checking if lastBlock is 0

				ef.log.Debug().Msgf("received pre-commit header %s", header.Number.String())

				if len(lastBlock.Bits()) == 0 { // This is the first header received
					lastBlock = header.Number
					color.Set(color.FgGreen)
					ef.log.Debug().Msgf("sending first block through channel, height: %s", header.Number.String())
					headerQueue <- header.Number
					color.Unset()
					notifierChannel <- true
				} else {
					// Set what we are expecting to receive
					expected := lastBlock.Add(lastBlock, big.NewInt(1))

					// Check to see if what we are receiving is what we are expecting
					if expected.Cmp(header.Number) == 0 {
						lastBlock = header.Number
						color.Set(color.FgCyan)
						ef.log.Debug().Msgf("sending block height through channel %s", header.Number.String())
						headerQueue <- header.Number
						color.Unset()
						notifierChannel <- true
					} else {
						// We need to recover the correct block here.
						// First, we need to find the difference between received and expected
						dif := big.NewInt(0)
						dif.Sub(expected, header.Number)

						// Check the difference between expected and received
						if expected.Cmp(header.Number) < 0 { // negative
							// If negative, then there are blocks that we missed

							//wg.Add(int(dif.Abs(dif).Int64()))

							// We can recover missed blocks here
							fmt.Println("Recovering blocks from ", expected.String(), " to ", header.Number.String())
							ef.log.Debug().Str("expected", expected.String()).Str("received", header.Number.String()).Msg("received block out of order")
							endloop := header.Number.Add(header.Number, big.NewInt(1))
							for i := new(big.Int).Set(expected); i.Cmp(endloop) < 0; i.Add(i, big.NewInt(1)) {
								x := new(big.Int).Set(i) //copy
								fmt.Println("recovering: ", x.String())
								headerQueue <- x
								notifierChannel <- true
							}
							//headerQueue <- header.Number
							//notifierChannel <- true
							lastBlock = header.Number

						} else { // positive
							// If positive, then the client node sent the same block header twice (because for some reason it does that)
							// Do nothing in this instance, since we already received it
							ef.log.Warn().Str("expected", expected.String()).Str("received", header.Number.String()).Msg("received block twice")
							// Decrement lastBlock by dif
							lastBlock.Sub(lastBlock, dif)
						}
					}
				}
			}
		}
	}()
	return retChan, nil
}*/

// This receives block headers and waits for them to reach a sufficient block height before pulling the data
/*func (ef *EventFeed) processBlockHeader(ch chan *big.Int, nch chan bool) chan *big.Int {

	retChan := make(chan *big.Int, config.Conf.ClientChain.MaxBufferSize)
	// Fire goroutine to listen
	go func() {
		for {
			// We will have a buffer channel that can store up to 50 block heights, and want to pull if > RequiredConfirmations
			if len(ch) >= config.Conf.ClientChain.RequiredConfirmations {
				// Once it is at 12, we can assume that this Ethereum fork is correct and pull the event data

				height := <-ch

				color.Set(color.FgYellow)
				ef.log.Debug().Msgf("block height %s has enough confirmations (%d needed, %d received)", height, config.Conf.ClientChain.RequiredConfirmations, len(ch)+1) // adding one since we just took one out
				color.Unset()
				fmt.Print("processing block height ", height.String(), "\n")
				retChan <- height
				<-nch
			} else {
				// Pull the notifier from the channel and restart the loop.
				// This is here so that it will hang until the header queue is long enough
				<-nch
				ef.log.Debug().Msgf("not enough block confirmations, adding block to queue")
			}
		}

	}()

	// Keeping this here for debugging diagnostic reasons.  If the above code breaks, the below code requires no block confirmations.
	/*go func() {
		for {
			h := <-ch
			color.Set(color.FgMagenta)
			fmt.Println("processing block height:", h)
			color.Unset()
			retChan <- h
			<-nch
		}
	}()

	return retChan
}*/
