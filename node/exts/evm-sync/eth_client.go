package evmsync

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jpillora/backoff"

	"github.com/kwilteam/kwil-db/core/log"
)

// ethClient is a client for interacting with the ethereum blockchain
// it handles retries and resubscribing to the blockchain in case of
// transient errors
type ethClient struct {
	targetAddress []ethcommon.Address
	topics        [][]ethcommon.Hash
	maxRetries    int64
	logger        log.Logger
	client        *ethclient.Client
	done          <-chan struct{}
}

// newEthClient creates a new ethereum client
func newEthClient(ctx context.Context, rpcurl string, maxRetries int64, targetAddresses []string, topics []string, done <-chan struct{}, logger log.Logger) (*ethClient, error) {
	var client *ethclient.Client

	addresses := make([]ethcommon.Address, len(targetAddresses))
	for i, addr := range targetAddresses {
		if !strings.HasPrefix(addr, "0x") {
			return nil, fmt.Errorf("invalid contract address: %s", addr)
		}
		if _, err := hex.DecodeString(addr[2:]); err != nil {
			return nil, fmt.Errorf("invalid contract address: %s", addr)
		}

		addresses[i] = ethcommon.HexToAddress(addr)
	}

	// default nil to allow all events
	var topicHashes []ethcommon.Hash
	for _, topic := range topics {
		// as per https://goethereumbook.org/event-read/#topics,
		// the first topic is the event signature
		topicHashes = append(topicHashes, crypto.Keccak256Hash([]byte(topic)))
	}

	// I don't set the max retries here because this only gets run on startup
	// the max retries are used for resubscribing to the blockchain
	// if we fail 3 times here, it is likely a permanent error
	count := 0
	err := retry(ctx, maxRetries, func() error {
		if count > 0 {
			logger.Warn("Retrying initial client connection", "attempt", count)
		}
		count++

		var innerErr error
		client, innerErr = ethclient.DialContext(ctx, rpcurl)
		return innerErr
	})
	if err != nil {
		return nil, err
	}

	return &ethClient{
		targetAddress: addresses,
		topics:        [][]ethcommon.Hash{topicHashes},
		maxRetries:    maxRetries,
		logger:        logger,
		client:        client,
		done:          done,
	}, nil
}

// GetLatestBlock gets the latest block number from the ethereum blockchain
func (ec *ethClient) GetLatestBlock(ctx context.Context) (int64, error) {
	var blockNumber int64
	err := retry(ctx, ec.maxRetries, func() error {
		block, err := ec.client.BlockNumber(ctx)
		if err != nil {
			ec.logger.Error("Failed to get latest block", "error", err)
			return err
		}
		blockNumber = int64(block)
		return nil
	})
	return blockNumber, err
}

// GetEventLogs gets the logs for the events from the ethereum blockchain.
// It can be given a start range and an end range to filter the logs by block height.
func (ec *ethClient) GetEventLogs(ctx context.Context, fromBlock, toBlock int64) ([]types.Log, error) {
	var logs []types.Log
	err := retry(ctx, ec.maxRetries, func() error {
		var err error
		logs, err = ec.client.FilterLogs(ctx, ethereum.FilterQuery{
			ToBlock:   big.NewInt(toBlock),
			FromBlock: big.NewInt(fromBlock),
			Addresses: ec.targetAddress,
			Topics:    ec.topics,
		})
		if err != nil {
			ec.logger.Error("Failed to get credit event logs", "error", err)
		}

		return err
	})
	return logs, err
}

// ListenToBlocks subscribes to new blocks on the ethereum blockchain.
// It takes a reconnectInterval, which is the amount of time it will wait
// to reconnect to the ethereum client if no new blocks are received.
// It takes a callback function that is called with the new block number.
// It can send duplicates, if that is received from the ethereum client.
// It will block until the context is cancelled, or until an error is
// returned from the callback function.
func (ec *ethClient) ListenToBlocks(ctx context.Context, reconnectInterval time.Duration, cb func(int64) error) error {
	headers := make(chan *types.Header)
	sub, err := ec.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return err
	}

	resubscribe := func() error {
		var retryCount int
		ec.logger.Warn("Resubscribing to Ethereum node", "attempt", retryCount) // anomalous
		sub.Unsubscribe()

		return retry(ctx, ec.maxRetries, func() error {
			retryCount++
			sub, err = ec.client.SubscribeNewHead(ctx, headers)
			return err
		})
	}

	reconn := time.NewTicker(reconnectInterval)
	defer reconn.Stop()

	for {
		select {
		case <-ctx.Done():
			ec.logger.Debug("Context cancelled, stopping ethereum client")
			return nil
		case <-ec.done:
			ec.logger.Debug("Done channel closed, stopping ethereum client")
			return nil
		case header := <-headers:
			ec.logger.Debug("New block", "height", header.Number.Int64())
			err := cb(header.Number.Int64())
			if err != nil {
				return err
			}

			reconn.Reset(reconnectInterval)
		case err := <-sub.Err():
			ec.logger.Error("Ethereum subscription error", "error", err)
			err = resubscribe()
			if err != nil {
				return err
			}
			reconn.Reset(reconnectInterval)
		case <-reconn.C:
			ec.logger.Warn("No new blocks received, resubscribing")
			err := resubscribe()
			if err != nil {
				return err
			}
		}
	}
}

// Close closes the ethereum client
func (ec *ethClient) Close() {
	ec.client.Close()
}

// retry will retry the function until it is successful, or reached the max retries
func retry(ctx context.Context, maxRetries int64, fn func() error) error {
	retrier := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	for {
		err := fn()
		if err == nil {
			return nil
		}

		// fail after maxRetries retries
		if retrier.Attempt() > float64(maxRetries) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retrier.Duration()):
		}
	}
}
