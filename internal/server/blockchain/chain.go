package blockchain

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kwilteam/kwil-db/internal/common/host"
	cfg "github.com/kwilteam/kwil-db/internal/common/host/config"
	types "github.com/kwilteam/kwil-db/internal/common/host/service"
)

const CHAIN_SERVICE_NAME string = "chain-service"

var chainIdentity types.ServiceIdentity

func init() {
	identity, err := host.RegisterService(CHAIN_SERVICE_NAME, func() types.Service {
		return &chainImpl{
			linger_ms: 100,
			quit:      make(chan struct{}),
			mu:        sync.Mutex{},
		}
	})
	if err != nil {
		panic(err)
	}
	chainIdentity = identity
}

// There will likely be multiple chains in the future, but will just assume a global default for now
func KwilMain() Chain {
	chain, err := host.GetServiceById[Chain](chainIdentity.Id())
	if err != nil {
		panic(err)
	}
	return chain
}

type chainImpl struct {
	types.BackgroundService
	handler   ChainTxHandler
	linger_ms int32
	height    uint64
	txns_chan chan *ChainTxCallback
	mu        sync.Mutex
	quit      chan struct{}
}

func (c *chainImpl) GetName() string {
	return "GLOBAL_CHAIN"
}

func (c *chainImpl) Configure(config cfg.Config) error {
	tx_max, err := config.GetInt32("block-tx-max", 1000)
	if err != nil {
		c.txns_chan = make(chan *ChainTxCallback, 1000)
		return err
	}

	linger_ms, err := config.GetInt32("linger_ms", 100)
	if err != nil {
		return err
	}

	c.linger_ms = linger_ms
	c.txns_chan = make(chan *ChainTxCallback, tx_max)

	return nil
}

func (c *chainImpl) Identity() types.ServiceIdentity {
	return chainIdentity
}

func (c *chainImpl) SetHandler(handler ChainTxHandler) {
	c.handler = handler //ToDo: check for nil and panic if already set
}

func (c *chainImpl) Submit(ctx context.Context, tx *BlockTx, cb func(context.Context, *BlockTx, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.IsRunning() {
		cb(ctx, nil, errors.New("chain already shutdown"))
		return
	}

	// NOTE - this is for dev end-to-end of the first cycle
	// The below only supports a single node or 10K tx.
	// Until we distributed rate limiting, the backpressure of
	// the channel is the only way to limit the number of txns

	// ToDo: create ChainContext and pass to ChainTxCallback.
	c.txns_chan <- &ChainTxCallback{nil, onCommitBlockResponse}
}

func onCommitBlockResponse(ctx ChainContext, e error) {
	fn := ctx.opaque().(func(context.Context, *BlockTx, error))
	if e != nil {
		fn(ctx, nil, e)
	} else {
		fn(ctx, ctx.tx(), nil)
	}
}

func (c *chainImpl) Start(ctx types.ServiceContext) error {
	// run at each interval of c.linger_ms
	ticker := time.NewTicker(time.Duration(c.linger_ms) * time.Millisecond)
	go func() {
		for {
			select {
			case <-c.quit:
				ticker.Stop()
				return
			case <-ticker.C:
				c.processTx(c.height)
			}

			if c.quit == nil {
				return
			}
		}
	}()

	return nil
}

func (c *chainImpl) processTx(height uint64) {
	for {
		select {
		case <-c.quit:
			return

		case cb, ok := <-c.txns_chan:
			if !ok {
				return
			}

			c.handler.CommitBlock(cb)
			if cb.ctx.GetHeight() >= height {
				return
			}
		}

		if c.quit == nil {
			return
		}
	}
}

func (c *chainImpl) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.handler != nil {
		c.handler.Close()
	}

	c.quit <- struct{}{}

	q := c.quit
	c.quit = nil
	close(q)

	t := c.txns_chan
	c.txns_chan = nil
	close(t)
}

func (c *chainImpl) IsRunning() bool {
	return c.quit != nil
}

func (c *chainImpl) AwaitShutdown() {
	<-c.quit
}
