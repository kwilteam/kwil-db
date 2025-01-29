package evmsync

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/extensions/listeners"
	"github.com/kwilteam/kwil-db/node/exts/evm-sync/chains"
	orderedsync "github.com/kwilteam/kwil-db/node/exts/ordered-sync"
)

func init() {
	err := listeners.RegisterListener("evm_sync", func(ctx context.Context, service *common.Service, eventstore listeners.EventStore) error {})
	if err != nil {
		panic(err)
	}
}

// syncConfig is a config that is shared by all listeners.
type syncConfig struct {
	// ReconnectionInterval is the amount of time in seconds that the listener
	// will wait before resubscribing for new Ethereum Blocks. Reconnects are
	// automatically handled, but a subscription may stall, in which case we
	// will make a new subscription. If the write or read on the connection to
	// the RPC provider errors, the RPC client will reconnect, and we will
	// continue to reestablish a new block subscription. If not configured, it
	// will default to 60s.
	ReconnectionInterval int64
	// MaxRetries is the total number of times the listener will attempt an RPC
	// with the provider before giving up. It will exponentially back off after
	// each try, starting at 1 second and doubling each time. If not configured,
	// it will default to 10.
	MaxRetries int64
}

func (c *syncConfig) load(m map[string]string) error {
	if v, ok := m["reconnection_interval"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		if i <= 0 {
			return errors.New("reconnection_interval must be greater than 0")
		}

		c.ReconnectionInterval = i
	} else {
		c.ReconnectionInterval = 60
	}

	if v, ok := m["max_retries"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		if i <= 0 {
			return errors.New("max_retries must be greater than 0")
		}

		c.MaxRetries = i
	} else {
		c.MaxRetries = 10
	}

	return nil
}

// chainConfig is a config that is specific to a single chain.
type chainConfig struct {
	// BlockSyncChunkSize is the number of Ethereum blocks the listener will request from the
	// Ethereum RPC endpoint at a time while catching up to the network. If not configured,
	// it will default to 1,000,000.
	BlockSyncChunkSize int64
	// Provider is the URL of the RPC endpoint for the chain.
	// It is required.
	Provider string
}

// load loads the config into the struct from the node's local configuration
func (c *chainConfig) load(m map[string]string) error {
	if v, ok := m["block_sync_chunk_size"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		if i <= 0 {
			return errors.New("block_sync_chunk_size must be greater than 0")
		}

		c.BlockSyncChunkSize = i
	} else {
		c.BlockSyncChunkSize = 1000000
	}

	v, ok := m["provider"]
	if !ok {
		return errors.New("provider is required")
	}

	c.Provider = v

	return nil
}

// individualListener is a singler configured client that is responsible for listening to a single set of contracts.
// Many individual listeners can exist for a single chain.
type individualListener struct {
	chain chains.ChainInfo
	// syncConf is the configuration for the listener.
	syncConf *syncConfig
	// chainConf is the configuration for the chain.
	chainConf *chainConfig
	// client is the Ethereum client that is used to listen to the chain.
	client *ethClient
	// orderedSyncTopic is the ordered sync topic that the listener is posting to.
	orderedSyncTopic string
}

func newIndividualListener(ctx context.Context, chain chains.ChainInfo, syncConf *syncConfig, chainConf *chainConfig,
	orderedSyncTopic string, contracts []string, topics []string, logger log.Logger) (*individualListener, error) {
	client, err := newEthClient(ctx, chainConf.Provider, syncConf.MaxRetries, contracts, topics, logger)
	if err != nil {
		return nil, err
	}

	return &individualListener{
		chain:            chain,
		syncConf:         syncConf,
		chainConf:        chainConf,
		client:           client,
		orderedSyncTopic: orderedSyncTopic,
	}, nil
}

// listen listens for new blocks from the Ethereum chain and broadcasts them to the network.
func (i *individualListener) listen(ctx context.Context, service *common.Service, eventstore listeners.EventStore) error {
	logger := service.Logger.New(i.orderedSyncTopic + "." + string(i.chain.Name))

	startBlock, err := getLastSeenHeight(ctx, eventstore, i.orderedSyncTopic)
	if err != nil {
		return fmt.Errorf("failed to get last seen height: %w", err)
	}

	currentBlock, err := i.client.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	if currentBlock < startBlock {
		return fmt.Errorf("starting height is greater than the last confirmed eth block height")
	}

	lastConfirmedBlock := currentBlock - i.chain.RequiredConfirmations
	logger.Infof(fmt.Sprintf("catching up from block %d to block %d", startBlock, lastConfirmedBlock))

	// we will now sync all logs from the starting height to the current height,
	// in chunks of config.BlockSyncChunkSize
	for {
		if startBlock >= lastConfirmedBlock {
			break
		}

		toBlock := startBlock + i.chainConf.BlockSyncChunkSize
		if toBlock > lastConfirmedBlock {
			toBlock = lastConfirmedBlock
		}

		err = i.processEvents(ctx, startBlock, toBlock, eventstore, logger)
		if err != nil {
			return err
		}

		startBlock = toBlock
	}

	logger.Info(fmt.Sprintf("synced up to block %d", lastConfirmedBlock))

	outerErr := i.client.ListenToBlocks(ctx, time.Duration(i.syncConf.ReconnectionInterval)*time.Second, func(newHeight int64) error {
		newHeight = newHeight - i.chain.RequiredConfirmations

		// it is possible to receive the same block height multiple times
		if newHeight <= startBlock {
			logger.Debug("received duplicate block height", "block", newHeight)
			return nil
		}

		logger.Info("received new block", "block", newHeight)

		// lastheight + 1 because we have already processed the last height
		err := i.processEvents(ctx, startBlock+1, newHeight, eventstore, logger)
		if err != nil {
			return err
		}

		startBlock = newHeight

		return nil
	})
	if outerErr != nil {
		return fmt.Errorf("failed to listen to blocks: %w", outerErr)
	}

	return nil
}

func (i *individualListener) processEvents(ctx context.Context, from, to int64, eventStore listeners.EventStore, logger log.Logger) error {
	logs, err := i.client.GetEventLogs(ctx, from, to)
	if err != nil {
		return err
	}

	blocks := make(map[uint64][]ethtypes.Log)
	blockOrder := make([]uint64, 0, len(blocks))

	for _, log := range logs {
		if len(log.Topics) == 0 {
			logger.Debug("skipping log without topics", "log", log)
			continue
		}

		_, ok := blocks[log.BlockNumber]
		if !ok {
			blockOrder = append(blockOrder, log.BlockNumber)
		}

		blocks[log.BlockNumber] = append(blocks[log.BlockNumber], log)
	}

	lastUsed, err := getLastHeightWithEvent(ctx, eventStore, i.orderedSyncTopic)
	if err != nil {
		return err
	}

	var lastUsedHeight *int64
	if lastUsed != -1 {
		lastUsedHeight2 := lastUsed
		lastUsedHeight = &lastUsedHeight2
	}

	for _, blockNum := range blockOrder {
		logs := blocks[blockNum]

		// ensure we are ordered by tx index
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].Index < logs[j].Index
		})

		serLogs, err := serializeLogs(logs)
		if err != nil {
			return fmt.Errorf("failed to serialize logs: %w", err)
		}

		data := orderedsync.ResolutionMessage{
			Topic:               i.orderedSyncTopic,
			PointInTime:         int64(blockNum),
			Data:                serLogs,
			PreviousPointInTime: lastUsedHeight,
		}

		bts, err := data.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal resolution message: %w", err)
		}

		err = eventStore.Broadcast(ctx, orderedsync.ExtensionName, bts)
		if err != nil {
			return fmt.Errorf("failed to broadcast resolution message: %w", err)
		}

		b2 := int64(blockNum)
		lastUsedHeight = &b2
	}

	// it is important that this is set before the last seen height is set, since it is ok
	// to resubmit the same event, but not okay to submit an event with an invalid last height
	if len(blockOrder) > 0 {
		err = setLastHeightWithEvent(ctx, eventStore, i.orderedSyncTopic, int64(blockOrder[len(blockOrder)-1]))
		if err != nil {
			return fmt.Errorf("failed to set last height with event: %w", err)
		}
	}

	logger.Info("processed events", "from", from, "to", to)

	return setLastSeenHeight(ctx, eventStore, i.orderedSyncTopic, to)
}

// serializeLogs serializes the logs into a byte slice.
func serializeLogs(logs []ethtypes.Log) ([]byte, error) {
	res := new(bytes.Buffer)
	for _, log := range logs {
		buf := new(bytes.Buffer)
		err := log.EncodeRLP(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to encode log: %w", err)
		}

		// write the length of the log
		err = binary.Write(res, binary.BigEndian, int32(buf.Len()))
		if err != nil {
			return nil, fmt.Errorf("failed to write log length: %w", err)
		}

		// write the log
		_, err = res.Write(buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("failed to write log: %w", err)
		}
	}

	return res.Bytes(), nil
}

// deserializeLogs deserializes the logs from a byte slice.
func deserializeLogs(bts []byte) ([]ethtypes.Log, error) {
	res := make([]ethtypes.Log, 0)
	buf := bytes.NewBuffer(bts)
	for buf.Len() > 0 {
		var logLen int32
		err := binary.Read(buf, binary.BigEndian, &logLen)
		if err != nil {
			return nil, fmt.Errorf("failed to read log length: %w", err)
		}

		logBts := buf.Next(int(logLen))
		log := ethtypes.Log{}
		err = rlp.DecodeBytes(logBts, &log)
		if err != nil {
			return nil, fmt.Errorf("failed to decode log: %w", err)
		}

		res = append(res, log)
	}

	return res, nil
}

var (
	// lastSeenHeightKey is the key used to store the last height processed by the listener
	lastSeenHeightKey = []byte("lh")
	// lastHeightWithEventKey is the key used to store the last height with an event processed by the listener
	lastHeightWithEventKey = []byte("lwe")
)

// getLastSeenHeight gets the last height seen on the Ethereum blockchain.
func getLastSeenHeight(ctx context.Context, eventStore listeners.EventStore, syncTopic string) (int64, error) {
	// get the last confirmed block height processed by the listener
	lastHeight, err := eventStore.Get(ctx, append(lastSeenHeightKey, []byte(syncTopic)...))
	if err != nil {
		return 0, fmt.Errorf("failed to get last seed block height: %w", err)
	}

	if len(lastHeight) == 0 {
		return 0, nil
	}

	return int64(binary.LittleEndian.Uint64(lastHeight)), nil
}

// setLastSeenHeight sets the last height seen on the Ethereum blockchain.
func setLastSeenHeight(ctx context.Context, eventStore listeners.EventStore, syncTopic string, height int64) error {
	heightBts := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBts, uint64(height))

	// set the last confirmed block height processed by the listener
	err := eventStore.Set(ctx, append(lastSeenHeightKey, []byte(syncTopic)...), heightBts)
	if err != nil {
		return fmt.Errorf("failed to set last seed block height: %w", err)
	}
	return nil
}

// getLastHeightWithEvent gets the last height that was processed by the listener and has an event.
// If this is the first one, it returns -1
func getLastHeightWithEvent(ctx context.Context, eventStore listeners.EventStore, syncTopic string) (int64, error) {
	lastHeight, err := eventStore.Get(ctx, append(lastHeightWithEventKey, []byte(syncTopic)...))
	if err != nil {
		return 0, fmt.Errorf("failed to get last used block height: %w", err)
	}

	if len(lastHeight) == 0 {
		return -1, nil
	}

	return int64(binary.LittleEndian.Uint64(lastHeight)), nil
}

// setLastHeightWithEvent sets the last height that was processed by the listener and has an event.
func setLastHeightWithEvent(ctx context.Context, eventStore listeners.EventStore, syncTopic string, height int64) error {
	heightBts := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBts, uint64(height))

	err := eventStore.Set(ctx, append(lastHeightWithEventKey, []byte(syncTopic)...), heightBts)
	if err != nil {
		return fmt.Errorf("failed to set last used block height: %w", err)
	}

	return nil
}
