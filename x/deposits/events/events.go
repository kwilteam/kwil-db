package events

import (
	"context"
	"errors"
	"time"

	"kwil/x/cfgx"
	cc "kwil/x/deposits/chainclient"
	"kwil/x/deposits/structures"
	ct "kwil/x/deposits/types"

	"kwil/x/logx"
)

type EventFeed interface {
	Listen(context.Context, int64) (<-chan int64, <-chan error, error)
	Contract() ct.Contract
	GetLastConfirmedBlock(context.Context) (int64, error)
}

type eventFeed struct {
	log      logx.SugaredLogger
	conf     cfgx.Config
	client   ct.Client
	reqConfs uint16
	timeout  time.Duration
	sc       ct.Contract
}

func New(c cfgx.Config, l logx.Logger) (EventFeed, error) {
	chnid := c.String("chain")
	logger := l.Sugar().With("chain", chnid)

	// build client
	cb := cc.Builder()

	client, err := cb.Logger(l).Chain(chnid).Endpoint(c.String("provider-endpoint")).Build()
	if err != nil {
		return nil, err
	}

	// get contract
	sc, err := client.GetContract(c.String("contract-address"))
	if err != nil {
		return nil, err
	}

	// get timeout
	toint64 := c.Int64("block-timeout", 30)

	// convert to duration
	timeout := time.Duration(toint64) * time.Second

	return &eventFeed{
		log:      logger,
		conf:     c,
		client:   client,
		reqConfs: uint16(c.Int64("required-confirmations", 12)), // 12 as default as defined by Ethereum yellow paper
		timeout:  timeout,
		sc:       sc,
	}, nil
}

// Listen listens for new block headers and sends them to the headers channel
func (ef *eventFeed) Listen(
	ctx context.Context, start int64,
) (<-chan int64, <-chan error, error) {
	ef.log.Info("starting event feed")

	return ef.listenForBlocks(ctx, start)
}

// I made Listen it's own function since we used to have extra logic here and might later

// ListenForBlocks listens for new block headers and sends them to the headers channel
// it should automatically reconnect if the connection is lost, and only send blocks that are finalized
func (ef *eventFeed) listenForBlocks(ctx context.Context, start int64) (<-chan int64, <-chan error, error) {
	headers := make(chan int64, 10) // this channel is for blocks before finalization
	errs := make(chan error)
	sub, err := ef.client.SubscribeBlocks(ctx, headers)
	if err != nil {
		return headers, errs, err
	}

	// adding buffers on retchan and headers to prevent blocking.
	// I had this in the config but took it out

	retChan := make(chan int64, 10) // getting returned, this is for blocks after finalization

	q := structures.NewQueue()

	go func(context.Context, int64) {
		exp := start
		for {
			select {
			case err := <-sub.Err():
				ef.log.Error("error from eth client, reconnecting", "err", err)
				sub.Unsubscribe()
				sub, err = ef.resubscribe(ctx, headers)
				if err != nil {
					ef.log.Error("error resubscribing to eth client, shutting down listener...", "err", err)
					errs <- err
					return
				}
			case <-ctx.Done():
				ef.log.Info("context done, shutting down listener...")
				sub.Unsubscribe()
				return
			case <-time.After(ef.timeout):
				ef.log.Warn("timeout waiting for block, reconnecting")
				sub.Unsubscribe()
				sub, err = ef.resubscribe(ctx, headers)
				if err != nil {
					ef.log.Error("error resubscribing to eth client, shutting down listener...", "err", err)
					errs <- err
					return
				}
			case header := <-headers:
				exp++ // expecting one more than the start / last block

				// now we need to ensure it is one greater than last
				if header == exp { // we got what we expected (1 more than last)
					q.Append(header)
				} else if header > exp { // received is greater than expected
					for i := exp; i <= header; i++ {
						q.Append(i)
					}
				}

				// now we set the last block to the tail
				exp = q.Tail()

				// now we need to check if we have any finalized blocks in the queue

				/*
					if q.Len() > ef.reqConfs {
						iters := q.Len() - ef.reqConfs
						for i := uint16(0); i < iters; i++ {
							retChan <- q.Pop()
						}
					}*/

				for {
					if q.Len() > ef.reqConfs {
						// we have enough blocks, send the oldest one
						retChan <- q.Pop()
					} else {
						// we don't have enough blocks, break out of the loop
						break
					}
				}

			}
		}
	}(ctx, start)

	return retChan, errs, nil
}

var ErrReconnect = errors.New("failed to reconnect")

// resubscribe will resubscribe to the block headers, and return the new subscription
func (ef *eventFeed) resubscribe(ctx context.Context, headers chan int64) (ct.BlockSubscription, error) {
	// I definitely need to refactor this
	ef.log.Debug("resubscribing to eth client")
	sub, err := ef.client.SubscribeBlocks(ctx, headers)
	if err != nil {
		ef.log.Warn("error resubscribing to eth client, retrying in 5 seconds...", "err", err)
		time.Sleep(time.Second)
		sub, err = ef.client.SubscribeBlocks(ctx, headers)
		if err != nil {
			ef.log.Warn("error resubscribing to eth client, retrying in 5 seconds...", "err", err)
			time.Sleep(5 * time.Second)
			sub, err = ef.client.SubscribeBlocks(ctx, headers)
			if err != nil {
				ef.log.Warn("error resubscribing to eth client, retrying in 1- seconds...", "err", err)
				time.Sleep(10 * time.Second)
				sub, err = ef.client.SubscribeBlocks(ctx, headers)
				if err != nil {
					ef.log.Warn("error resubscribing to eth client, shutting down event listener...", "err", err)
					return nil, ErrReconnect
				}
			}
		}
	}

	return sub, nil
}

func (ef *eventFeed) Contract() ct.Contract {
	return ef.sc
}

func (ef *eventFeed) GetLastConfirmedBlock(ctx context.Context) (int64, error) {
	// get the most recent block from ethereum and subtract reqConfs
	lb, err := ef.client.GetLatestBlock(ctx)
	if err != nil {
		return 0, err
	}

	return lb - int64(ef.reqConfs), nil
}
