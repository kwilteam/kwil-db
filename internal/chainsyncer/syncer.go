package chainsyncer

import (
	"context"
	"fmt"

	bClient "github.com/kwilteam/kwil-db/core/bridge/client"
	syncer "github.com/kwilteam/kwil-db/core/bridge/syncer"
	"github.com/kwilteam/kwil-db/core/types/chain"
	"go.uber.org/zap"
)

/*
	Chain Syncer:
		- Listens for the deposit events and saves them to the database
		- Requirements:
			- BridgeClient (interface): To interact with the smart contracts
			- BlockSyncer (interface): To listen for the new blocks (confirmed)
*/

type ChainSyncer struct {
	bridgeClient bClient.BridgeClient
	blockSyncer  syncer.BlockSyncer
	// eventStore  EventStore

	// height of the last block that was synced
	height int64

	// chunk size
	chunkSize int64
}

func New(bridgeClient bClient.BridgeClient, blockSyncer syncer.BlockSyncer) *ChainSyncer {
	return &ChainSyncer{
		bridgeClient: bridgeClient,
		blockSyncer:  blockSyncer,
		// eventStore:  eventStore,
		height:    0,
		chunkSize: 100,
	}
}

func (cs *ChainSyncer) Start() error {
	ctx := context.Background()

	latestHeight, err := cs.blockSyncer.LatestBlock(ctx)
	if err != nil {
		return err
	}
	fmt.Println("latestHeight", latestHeight)

	// retrieve the deposits from the last synced block to the latest block
	return cs.listen(ctx)
}

// listen listens for new blocks on the chain and syncs them as they come in
func (cs *ChainSyncer) listen(ctx context.Context) error {
	blockChan := make(chan int64)
	err := cs.blockSyncer.Listen(ctx, blockChan)
	if err != nil {
		return err
	}

	go func(blockChan <-chan int64) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Chain syncer panic", zap.Any("error", err))
			}
			fmt.Println("Chain syncer stopped")
		}()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stop Chain syncer")
				return
			case block := <-blockChan:
				fmt.Println("Received block", zap.Int64("block", block))

				depositEvents, err := cs.getDepositEvents(ctx, block, block)
				if err != nil {
					fmt.Println("Failed to get deposit events for block", zap.Int64("block", block), zap.Error(err))
					return
				}

				for _, depositEvent := range depositEvents {
					fmt.Println("depositEvent", depositEvent)
				}
			}
		}
	}(blockChan)

	return nil
}

func (cs *ChainSyncer) getDepositEvents(ctx context.Context, from int64, to int64) ([]*chain.DepositEvent, error) {
	ctr := cs.bridgeClient.EscrowContract()
	end := uint64(to)
	depositEvents, err := ctr.GetDeposits(ctx, uint64(from), &end)
	if err != nil {
		return nil, err
	}

	return depositEvents, nil
}

func (cs *ChainSyncer) Close() error {
	return cs.blockSyncer.Close()
}
