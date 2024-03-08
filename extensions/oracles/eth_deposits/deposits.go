// package ethdeposits implements an oracle that listens to Ethereum events
// and triggers the creation of deposit events in Kwil.
package ethdeposits

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/extensions/oracles"
	"github.com/kwilteam/kwil-db/extensions/resolutions/credit"
)

const OracleName = "eth_deposits"

// use golang's init function, which runs before main, to register the extension
// see more here: https://www.digitalocean.com/community/tutorials/understanding-init-in-go
func init() {
	// register the oracle with the name "eth_deposit"
	err := oracles.RegisterOracle(OracleName, Start)
	if err != nil {
		panic(err)
	}
}

// Start starts the eth_deposit oracle, which triggers the creation of deposit events in Kwil.
// It can be configured to listen to a certain smart contract address. It will listen for the EVM event signature
// "Credit(address,uint256)" and create a deposit event in Kwil when it sees a matching event. It uses the
// "credit_account" resolution, defined in extensions/resolutions/credit/credit.go, to create the deposit event.
// It will search for a local extension configuration named "eth_deposit".
func Start(ctx context.Context, service *common.Service, eventstore oracles.EventStore) error {
	config := &EthDepositConfig{}
	oracleConfig, ok := service.ExtensionConfigs[OracleName]
	if !ok {
		service.Logger.Info("no eth_deposit configuration found, eth_deposit oracle will not start")
		return nil // no configuration, so we don't start the oracle
	}
	err := config.setConfig(oracleConfig)
	if err != nil {
		return fmt.Errorf("failed to set eth_deposit configuration: %w", err)
	}

	// we need to catch up with the ethereum chain.
	// we will get the last seen height from the kv store
	// we will either start from the last seen height, or from the configured starting height,
	// whichever is greater
	lastHeight, err := getLastStoredHeight(ctx, eventstore)
	if err != nil {
		return fmt.Errorf("failed to get last stored height: %w", err)
	}

	if config.StartingHeight > lastHeight {
		lastHeight = config.StartingHeight
	}

	client, err := newEthClient(ctx, config.RPCProvider, config.MaxRetries, ethcommon.HexToAddress(config.ContractAddress), service.Logger)
	if err != nil {
		return fmt.Errorf("failed to create ethereum client: %w", err)
	}
	defer client.Close()

	// get the current block height from the Ethereum client
	currentHeight, err := client.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block height: %w", err)
	}

	if lastHeight > currentHeight-config.RequiredConfirmations {
		return fmt.Errorf("starting height is greater than the last confirmed eth block height")
	}

	// we will now sync all logs from the starting height to the current height,
	// in chunks of config.BlockSyncChunkSize
	for {
		if lastHeight >= currentHeight-config.RequiredConfirmations {
			break
		}

		// get the next block chunk. if it is greater than the current height - required confirmations,
		// we will set it to the current height - required confirmations
		toBlock := lastHeight + config.BlockSyncChunkSize
		if toBlock > currentHeight-config.RequiredConfirmations {
			toBlock = currentHeight - config.RequiredConfirmations
		}

		err = processEvents(ctx, lastHeight, toBlock, client, eventstore, service.Logger)
		if err != nil {
			return fmt.Errorf("failed to process events: %w", err)
		}

		lastHeight = toBlock
	}

	// ListenToBlocks will listen to new blocks and process the events.
	// It only returns when the context is cancelled, or when the client cannot recover
	// from an error after the max retries.
	outerErr := client.ListenToBlocks(ctx, time.Duration(config.ReconnectionInterval)*time.Second, func(newHeight int64) error {
		newHeight = newHeight - config.RequiredConfirmations // account for required confirmations

		// it is possible to receive the same height twice
		if newHeight <= lastHeight {
			service.Logger.Info("received duplicate block height", "height", newHeight)
			return nil
		}

		service.Logger.Info("received new block height", "height", newHeight)

		// lastheight + 1 because we have already processed the last height
		err = processEvents(ctx, lastHeight+1, newHeight, client, eventstore, service.Logger)
		if err != nil {
			return fmt.Errorf("failed to process events: %w", err)
		}

		lastHeight = newHeight

		return nil
	})
	if outerErr != nil {
		return fmt.Errorf("ethereum client error: %w", outerErr)
	}

	<-ctx.Done() // wait for the context to be cancelled
	return nil
}

// processEvents will process all events from the Ethereum client from the given height range.
func processEvents(ctx context.Context, from, to int64, client *ethClient, eventstore oracles.EventStore, logger log.SugaredLogger) error {
	logs, err := client.GetCreditEventLogs(ctx, from, to)
	if err != nil {
		return fmt.Errorf("failed to get credit event logs: %w", err)
	}

	for _, log := range logs {
		event, err := decodeCreditEvent(&log)
		if err != nil {
			return fmt.Errorf("failed to decode credit event: %w", err)
		}

		bts, err := event.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		err = eventstore.Broadcast(ctx, credit.CreditAccountEventType, bts)
		if err != nil {
			return fmt.Errorf("failed to broadcast event: %w", err)
		}
	}

	logger.Info("processed events", "from", from, "to", to, "events", len(logs))

	return setLastStoredHeight(ctx, eventstore, to)
}

// EthDepositConfig is the configuration for the eth_deposit oracle.
// It can be read in from a map[string]string, which is passed from
// the node's local configuration.
type EthDepositConfig struct {
	// StartingHeight is the Ethereum block height it will start listening from.
	// Any events emitted before this height will be ignored.
	// If not configured, it will start from block 0.
	StartingHeight int64
	// ContractAddress is the Ethereum address of the smart contract it will listen to.
	// It is a required configuration.
	ContractAddress string
	// RequiredConfirmations is the number of Ethereum blocks that must be mined before
	// the oracle will create a deposit event in Kwil. This is to protect against Ethereum
	// network reorgs / soft forks. If not configured, it will default to 12.
	// https://www.alchemy.com/overviews/what-is-a-reorg
	RequiredConfirmations int64
	// RPCProvider is the URL of the Ethereum RPC endpoint it will connect to.
	// This would likely be an Infura / Alchemy endpoint.
	// It is a required configuration.
	RPCProvider string
	// ReconnectionInterval is the amount of time in seconds that the oracle will wait
	// before reconnecting to the Ethereum RPC endpoint if it is disconnected. Long-running
	// RPC subscriptions are prone to being reset by the Ethereum RPC endpoint, so this
	// will allow the oracle to reconnect. If not configured, it will default to 60s.
	ReconnectionInterval int64
	// MaxRetries is the total number of times the oracle will attempt to reconnect to the
	// Ethereum RPC endpoint before giving up. It will exponentially back off after each try,
	// starting at 1 second and doubling each time.
	// If not configured, it will default to 10.
	MaxRetries int64
	// BlockSyncChunkSize is the number of Ethereum blocks the oracle will request from the
	// Ethereum RPC endpoint at a time while catching up to the network. If not configured,
	// it will default to 1,000,000.
	BlockSyncChunkSize int64
}

// setConfig sets the configuration for the eth_deposit oracle.
// If it doesn't find a required configuration, or if it finds an invalid
// configuration, it returns an error
func (e *EthDepositConfig) setConfig(m map[string]string) error {
	startingHeight, ok := m["starting_height"]
	if !ok {
		startingHeight = "0"
	}

	var err error
	e.StartingHeight, err = strconv.ParseInt(startingHeight, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid starting_height: %s", startingHeight)
	}
	if e.StartingHeight < 0 {
		return fmt.Errorf("starting_height cannot be negative")
	}

	contractAddress, ok := m["contract_address"]
	if !ok {
		return fmt.Errorf("no contract_address provided")
	}
	e.ContractAddress = contractAddress

	requiredConfirmations, ok := m["required_confirmations"]
	if !ok {
		requiredConfirmations = "12"
	}
	e.RequiredConfirmations, err = strconv.ParseInt(requiredConfirmations, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid required_confirmations: %s", requiredConfirmations)
	}
	if e.RequiredConfirmations < 0 {
		return fmt.Errorf("required_confirmations cannot be negative")
	}

	rpc, ok := m["rpc_provider"]
	if !ok {
		return fmt.Errorf("no rpc_provider provided")
	}
	if !strings.HasPrefix(rpc, "ws") {
		return fmt.Errorf("rpc_provider must be a websocket URL")
	}
	e.RPCProvider = rpc

	reconnectionInterval, ok := m["reconnection_interval"]
	if !ok {
		reconnectionInterval = "60"
	}
	intervalInt, err := strconv.ParseInt(reconnectionInterval, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid reconnection_interval: %s", reconnectionInterval)
	}
	if intervalInt < 5 {
		return fmt.Errorf("reconnection_interval must be greater than or equal to 5")
	}
	e.ReconnectionInterval = intervalInt

	maxRetries, ok := m["max_retries"]
	if !ok {
		maxRetries = "10"
	}
	e.MaxRetries, err = strconv.ParseInt(maxRetries, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid max_retries: %s", maxRetries)
	}
	if e.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	blockSyncChunkSize, ok := m["block_sync_chunk_size"]
	if !ok {
		blockSyncChunkSize = "1000000"
	}
	e.BlockSyncChunkSize, err = strconv.ParseInt(blockSyncChunkSize, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid block_sync_chunk_size: %s", blockSyncChunkSize)
	}
	if e.BlockSyncChunkSize <= 0 {
		return fmt.Errorf("block_sync_chunk_size must be greater than 0")
	}

	return nil
}

// Map returns the configuration as a map[string]string.
// This is used for testing
func (e *EthDepositConfig) Map() map[string]string {

	return map[string]string{
		"starting_height":        strconv.FormatInt(e.StartingHeight, 10),
		"contract_address":       e.ContractAddress,
		"required_confirmations": strconv.FormatInt(e.RequiredConfirmations, 10),
		"rpc_provider":           e.RPCProvider,
		"reconnection_interval":  strconv.FormatInt(e.ReconnectionInterval, 10),
		"max_retries":            strconv.FormatInt(e.MaxRetries, 10),
		"block_sync_chunk_size":  strconv.FormatInt(e.BlockSyncChunkSize, 10),
	}
}

var (
	// lastHeightKey is the key used to store the last height processed by the oracle
	lastHeightKey = []byte("lh")
)

// getLastStoredHeight gets the last height stored by the KV store
func getLastStoredHeight(ctx context.Context, eventstore oracles.EventStore) (int64, error) {
	// get the last confirmed block height processed by the oracle
	lastHeight, err := eventstore.Get(ctx, lastHeightKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get last block height: %w", err)
	}

	if len(lastHeight) == 0 {
		return 0, nil
	}

	return int64(binary.LittleEndian.Uint64(lastHeight)), nil
}

// setLastStoredHeight sets the last height stored by the KV store
func setLastStoredHeight(ctx context.Context, eventstore oracles.EventStore, height int64) error {
	heightBts := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBts, uint64(height))

	// set the last confirmed block height processed by the oracle
	err := eventstore.Set(ctx, lastHeightKey, heightBts)
	if err != nil {
		return fmt.Errorf("failed to set last block height: %w", err)
	}
	return nil
}
