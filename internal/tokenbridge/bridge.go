package tokenbridge

import (
	"context"
	"fmt"
	"time"

	bClient "github.com/kwilteam/kwil-db/core/bridge/client"
	syncer "github.com/kwilteam/kwil-db/core/bridge/syncer"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/chain"
	"go.uber.org/zap"
)

/*
	TokenBridge:
		- Listens for the deposit events and saves them to the event store
		- Requirements:
			- TokenBridgeClient (interface): To interact with the smart contracts
			- BlockSyncer (interface): To listen for the new blocks (confirmed)
*/

type TokenBridge struct {
	bridgeClient   bClient.TokenBridgeClient
	blockSyncer    syncer.BlockSyncer
	depositStore   *DepositStore /// SHould we define an interface for this?
	startingHeight int64
	chunkSize      int64
	nodeAddress    string
	log            log.Logger
}

func New(bridgeClient bClient.TokenBridgeClient, blockSyncer syncer.BlockSyncer, depositStore *DepositStore, opts ...TokenBridgeOpts) *TokenBridge {
	tb := &TokenBridge{
		bridgeClient:   bridgeClient,
		blockSyncer:    blockSyncer,
		depositStore:   depositStore,
		chunkSize:      10000,
		startingHeight: 0,
		log:            log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(tb)
	}
	return tb
}

func (cs *TokenBridge) Start() error {
	ctx := context.Background()

	// get the last processed block
	startHeight, err := cs.depositStore.LastProcessedBlock(ctx)
	if err != nil {
		return err
	}
	startHeight = max(startHeight, cs.startingHeight)

	// last finalized block in the chain
	latestHeight, err := cs.blockSyncer.LatestBlock(ctx)
	if err != nil {
		return err
	}

	// retrieve all the deposits from the last processed block to the latest block
	for i := startHeight; i < latestHeight.Height; i += cs.chunkSize {
		end := min(i+cs.chunkSize-1, latestHeight.Height)
		err := cs.syncDepositEventsForRange(ctx, i, end)
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond) // chains won't like it if we hit them too fast
	}

	// retrieve the deposits from the last synced block to the latest block
	return cs.listen(ctx)
}

// listen listens for new blocks on the chain and syncs them as they come in
func (tb *TokenBridge) listen(ctx context.Context) error {
	blockChan := make(chan int64)
	err := tb.blockSyncer.Listen(ctx, blockChan)
	if err != nil {
		return err
	}

	go func(blockChan <-chan int64) {
		defer func() {
			if err := recover(); err != nil {
				tb.log.Error("TokenBridge panic", zap.Any("error", err))
			}
			tb.log.Info("TokenBridge stopped")
		}()
		for {
			select {
			case <-ctx.Done():
				tb.log.Info("TokenBridge stopped")
				return
			case block := <-blockChan:
				tb.log.Info("Received block", zap.Int64("block", block))

				err := tb.syncDepositEventsForRange(ctx, block, block)
				if err != nil {
					fmt.Println("Failed to sync deposit events for block", zap.Int64("block", block), zap.Error(err))
					return
				}

			}
		}
	}(blockChan)

	return nil
}

func (cs *TokenBridge) getDepositEvents(ctx context.Context, from int64, to int64) ([]*chain.DepositEvent, error) {
	end := uint64(to)
	depositEvents, err := cs.bridgeClient.GetDeposits(ctx, uint64(from), &end)
	if err != nil {
		return nil, err
	}

	return depositEvents, nil
}

func (cs *TokenBridge) Close() error {
	return cs.blockSyncer.Close()
}

func (tb *TokenBridge) syncDepositEventsForRange(ctx context.Context, from int64, to int64) error {
	depositEvents, err := tb.getDepositEvents(ctx, from, to)
	if err != nil {
		tb.log.Error("Failed to get deposit events for block range", zap.Int64("from", from), zap.Int64("to", to), zap.Error(err))
		return err
	}

	for _, depositEvent := range depositEvents {
		// Insert local event
		// TODO: Keep the amount as big.Int throughout the code
		err = tb.depositStore.AddDeposit(ctx, depositEvent.ID, depositEvent.Sender, depositEvent.Amount, tb.nodeAddress)
		if err != nil {
			tb.log.Error("Failed to add local event", zap.Error(err))
			return err
		}

		err = tb.depositStore.SetLastProcessedBlock(ctx, to)
		if err != nil {
			tb.log.Error("Failed to set last processed block", zap.Error(err))
			return err
		}
		tb.log.Info("depositEvent added to the eventstore: ", zap.Any("depositEvent", depositEvent))
	}
	return nil
}
