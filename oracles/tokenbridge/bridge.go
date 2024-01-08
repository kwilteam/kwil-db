package tokenbridge

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	ctypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/chain"
	"github.com/kwilteam/kwil-db/internal/kv"
	"github.com/kwilteam/kwil-db/oracles"
	bClient "github.com/kwilteam/kwil-db/oracles/tokenbridge/client"
	syncer "github.com/kwilteam/kwil-db/oracles/tokenbridge/syncer"
	"github.com/kwilteam/kwil-db/oracles/tokenbridge/types"
	"go.uber.org/zap"
)

const (
	oracleName           = "token_bridge"
	last_processed_block = "last_processed_block"
)

type TokenBridge struct {
	cfg          TokenBridgeConfig
	bridgeClient bClient.TokenBridgeClient
	blockSyncer  syncer.BlockSyncer

	datastore  ctypes.Datastores
	eventstore ctypes.EventStore
	kvstore    ctypes.KVStore

	log log.Logger
}

type TokenBridgeConfig struct {
	endpoint              string
	escrowAddress         string
	chainCode             chain.ChainCode
	startingHeight        int64
	chunkSize             int64
	requiredConfirmations int64
	ReconnectInterval     time.Duration
}

func init() {
	fmt.Println("Registering oracle", oracleName)
	tb := &TokenBridge{}
	err := oracles.RegisterOracle(oracleName, tb)
	if err != nil {
		fmt.Println("Failed to register oracle", zap.Error(err))
		panic(err)
	}

	payload := &types.AccountCredit{}
	err = ctypes.RegisterPaylod(payload)
	if err != nil {
		fmt.Println("Failed to register payload", zap.Error(err))
		panic(err)
	}
}

func (tb *TokenBridge) Name() string {
	return oracleName
}

func (cs *TokenBridge) Stop() error {
	if cs == nil {
		return nil
	}
	return cs.blockSyncer.Close()
}

func (tb *TokenBridge) Start(ctx context.Context, datastores ctypes.Datastores, eventstore ctypes.EventStore, logger log.Logger, metadata map[string]string) error {
	err := tb.Initialize(ctx, datastores, eventstore, logger, metadata)
	if err != nil {
		return err
	}

	// get the last processed block
	startHeight, err := tb.getBlockHeight(ctx)
	if err != nil {
		tb.log.Error("Failed to get last processed block", zap.Error(err))
		return err
	}

	startHeight = max(startHeight, tb.cfg.startingHeight)

	// last finalized block in the chain
	// TODO: FIX IT: Umm, should we error out if the chain doesn't have any finalized blocks? - probably not
	latestHeight, err := tb.blockSyncer.LatestBlock(ctx)
	if err != nil {
		return err
	}

	// retrieve all the deposits from the last processed block to the latest block
	for i := startHeight; i < latestHeight.Height; i += tb.cfg.chunkSize {
		end := min(i+tb.cfg.chunkSize-1, latestHeight.Height)
		err := tb.syncDepositEventsForRange(ctx, i, end)
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond) // chains won't like it if we hit them too fast
	}

	// retrieve the deposits from the last synced block to the latest block
	return tb.listen(ctx)
}

func (tb *TokenBridge) Initialize(ctx context.Context, datastores ctypes.Datastores, eventstore ctypes.EventStore, logger log.Logger, metadata map[string]string) error {
	tb.log = logger
	tb.log.Info("Initializing TokenBridge")
	tb.datastore = datastores
	tb.eventstore = eventstore
	tb.kvstore = eventstore.KV([]byte(oracleName))

	// Extract config from metadata
	err := tb.extractConfig(metadata)
	if err != nil {
		return err
	}

	// Initialize BridgeClient
	bridgeClient, err := bClient.New(ctx, tb.cfg.endpoint, tb.cfg.chainCode, tb.cfg.escrowAddress)
	if err != nil {
		return err
	}
	tb.bridgeClient = bridgeClient

	// Initialize BlockSyncer
	// TODO: How to pass logger??
	blockSyncer, err := syncer.New(bridgeClient, syncer.WithRequiredConfirmations(tb.cfg.requiredConfirmations))
	if err != nil {
		return err
	}
	tb.blockSyncer = blockSyncer

	return nil
}

func (tb *TokenBridge) extractConfig(metadata map[string]string) error {
	// Endpoint, EscrowAddress, ChainCode
	if endpoint, ok := metadata["endpoint"]; ok {
		tb.cfg.endpoint = endpoint
	} else {
		return fmt.Errorf("no endpoint provided")
	}

	if escrowAddr, ok := metadata["escrow_address"]; ok {
		tb.cfg.escrowAddress = escrowAddr
	} else {
		return fmt.Errorf("no escrow address provided")
	}

	if code, ok := metadata["chain_code"]; ok {
		// convert code to int64
		code64, err := strconv.ParseInt(code, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse chain code: %w", err)
		}
		tb.cfg.chainCode = chain.ChainCode(code64)
	}

	if confirmations, ok := metadata["required_confirmations"]; ok {
		// convert confirmations to int64
		confirmations64, err := strconv.ParseInt(confirmations, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse required confirmations: %w", err)
		}
		tb.cfg.requiredConfirmations = confirmations64
	} else {
		tb.cfg.requiredConfirmations = 12
	}

	if startingHeight, ok := metadata["starting_height"]; ok {
		// convert startingHeight to int64
		startingHeight64, err := strconv.ParseInt(startingHeight, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse starting height: %w", err)
		}
		tb.cfg.startingHeight = startingHeight64
	} else {
		tb.cfg.startingHeight = 0
	}

	// if chunkSize, ok := metadata["chunk_size"].(int64); ok {
	// 	tb.cfg.chunkSize = chunkSize
	// }
	tb.cfg.chunkSize = 10000

	if interval, ok := metadata["reconnect_interval"]; ok {
		// convert interval to float64
		interval64, err := strconv.ParseFloat(interval, 64)
		if err != nil {
			return fmt.Errorf("failed to parse reconnect interval: %w", err)
		}
		tb.cfg.ReconnectInterval = time.Duration(interval64) * time.Second
	}

	return nil
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
				tb.log.Info("Received Finalized block", zap.Int64("block", block))

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

func (cs *TokenBridge) getDepositEvents(ctx context.Context, from int64, to int64) ([]*types.AccountCredit, error) {
	end := uint64(to)
	depositEvents, err := cs.bridgeClient.GetDeposits(ctx, uint64(from), &end)
	if err != nil {
		return nil, err
	}

	return depositEvents, nil
}

func (tb *TokenBridge) syncDepositEventsForRange(ctx context.Context, from int64, to int64) error {
	tb.log.Info("Sync Deposits for range", zap.Int64("from", from), zap.Int64("to", to))
	depositEvents, err := tb.getDepositEvents(ctx, from, to)
	if err != nil {
		tb.log.Error("Failed to get deposit events for block range", zap.Int64("from", from), zap.Int64("to", to), zap.Error(err))
		return err
	}

	for _, depositEvent := range depositEvents {
		// Insert local event
		tb.log.Info("Adding deposit event to the eventstore", zap.Any("depositEvent", depositEvent))

		// Add deposit event to the event store
		bts, err := depositEvent.MarshalBinary()
		if err != nil {
			tb.log.Error("Failed to marshal deposit event", zap.Error(err))
			return err
		}

		event := &ctypes.Event{
			EventType: depositEvent.Type(),
			Data:      bts,
		}

		err = tb.eventstore.Store(ctx, event.Data, event.EventType)
		if err != nil {
			tb.log.Error("Failed to store event", zap.Error(err))
			return err
		}

		err = tb.setBlockHeight(ctx, to)
		if err != nil {
			tb.log.Error("Failed to set last processed block", zap.Error(err))
			return err
		}

		tb.log.Info("depositEvent pushed on the event channel: ", zap.Any("depositEvent", depositEvent))

	}
	return nil
}

func (tb *TokenBridge) getBlockHeight(ctx context.Context) (int64, error) {
	blockBytes, err := tb.kvstore.Get(ctx, []byte(last_processed_block))
	if err == kv.ErrKeyNotFound || blockBytes == nil {
		tb.setBlockHeight(ctx, 0)
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	height := binary.BigEndian.Uint64(blockBytes)
	fmt.Println("Last processed block", height)
	return int64(height), nil
}

func (tb *TokenBridge) setBlockHeight(ctx context.Context, height int64) error {
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(height))
	return tb.kvstore.Set(ctx, []byte(last_processed_block), heightBytes)
}

func (tb *TokenBridge) Events(ctx context.Context) ([]*ctypes.VotableEvent, error) {
	return tb.eventstore.GetEvents(ctx)
}
