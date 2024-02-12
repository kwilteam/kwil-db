package deposit_oracle

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jpillora/backoff"
	"github.com/kwilteam/kwil-db/internal/kv"
	"go.uber.org/zap"
)

func (do *EthDepositOracle) listen(ctx context.Context) error {
	// Listen for new blocks
	headers := make(chan *types.Header, 1)
	sub, err := do.ethclient.SubscribeNewHead(ctx, headers)
	if err != nil {
		return err
	}

	go func(ctx context.Context, sub ethereum.Subscription) {
		defer sub.Unsubscribe()
		requiredConfirmations := do.cfg.requiredConfirmations

		// catchup with the ethereum chain
		lastHeight, err := do.catchupWithEthChain(ctx)
		if err != nil {
			do.logger.Error("Failed to catchup with ethereum chain", zap.Error(err))
			return
		}

		do.logger.Info("Started listening for new blocks on ethereum: ", zap.Int64("lastHeight", lastHeight), zap.Int64("requiredConfirmations", requiredConfirmations))

		for {
			select {
			case <-ctx.Done():
				return
			case <-do.done:
				return
			case err := <-sub.Err():
				if err != nil {
					do.logger.Warn("subscription error", zap.Error(err))
					sub, err = do.resubscribe(ctx, sub, headers, do.cfg.maxRetries)
					if err != nil {
						do.logger.Error("Failed to resubscribe", zap.Error(err))
						return
					}
				}
			case header := <-headers:
				currentHeight := header.Number.Int64()
				do.logger.Info("New block", zap.Int64("height", currentHeight))

				if currentHeight-lastHeight < requiredConfirmations {
					continue
				}

				// get all the deposit events
				FromBlock := lastHeight + 1
				ToBlock := currentHeight - requiredConfirmations
				events, err := do.filterLogs(ctx, FromBlock, ToBlock)
				if err != nil {
					do.logger.Error("Failed to filter logs", zap.Error(err), zap.Int64("from", FromBlock), zap.Int64("to", ToBlock))
					return
				}

				for _, event := range events {
					do.addEvent(ctx, event)
				}
				lastHeight = currentHeight - requiredConfirmations
				do.setBlockHeight(ctx, lastHeight)
			}
		}
	}(ctx, sub)

	return nil
}

// catchupWithEthChain catches up with the ethereum chain by fetching all the deposit events
// from the last processed block to the latest block and recording them in the eventstore
// It returns the latest block height processed
func (do *EthDepositOracle) catchupWithEthChain(ctx context.Context) (int64, error) {
	// get the latest block on ethereum
	latestBlock, err := do.ethclient.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}

	startHeight, err := do.getBlockHeight(ctx)
	if err != nil {
		do.logger.Error("Failed to get last processed block", zap.Error(err))
		return 0, err
	}
	startHeight = max(startHeight, do.cfg.startingHeight)

	// get all the deposit events
	FromBlock := startHeight
	ToBlock := latestBlock.Number().Int64() - do.cfg.requiredConfirmations
	events, err := do.filterLogs(ctx, FromBlock, ToBlock)
	if err != nil {
		return 0, err
	}

	for _, event := range events {
		do.addEvent(ctx, event)
	}
	do.setBlockHeight(ctx, ToBlock)

	return ToBlock, nil
}

func (do *EthDepositOracle) resubscribe(ctx context.Context, sub ethereum.Subscription, headers chan *types.Header, maxRetries uint64) (ethereum.Subscription, error) {
	sub.Unsubscribe()

	retrier := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: true,
	}
	// keep trying to resubscribe until it works
	for {
		do.logger.Debug("Resubscribing to the ethereum node ", zap.Float64("attempt", retrier.Attempt()), zap.Uint64("maxRetries", maxRetries))
		sub, err := do.ethclient.SubscribeNewHead(ctx, headers)
		if err != nil {
			// fail after 50 retries,
			// TODO: shld we make this configurable
			if retrier.Attempt() > float64(maxRetries) {
				return nil, err
			}

			time.Sleep(retrier.Duration())
			continue
		}
		retrier.Reset()
		return sub, nil
	}
}

func (do *EthDepositOracle) filterLogs(ctx context.Context, from int64, to int64) ([]*AccountCredit, error) {
	// Make the queries in batches of do.cfg.maxTotalRequests to avoid overloading the server
	events := make([]*AccountCredit, 0)
	for i := from; i <= to; i += do.cfg.maxTotalRequests {
		endBlock := min(i+do.cfg.maxTotalRequests, to)
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(i),
			ToBlock:   big.NewInt(endBlock),
			Addresses: []common.Address{common.HexToAddress(do.cfg.escrowAddress)},
			Topics:    [][]common.Hash{{do.creditEventSignature}},
		}

		// retry the query until it works
		retrier := &backoff.Backoff{
			Min:    1 * time.Second,
			Max:    10 * time.Second,
			Factor: 2,
			Jitter: true,
		}
		for {
			logs, err := do.ethclient.FilterLogs(ctx, query)
			if err != nil {
				// fail after 15 retries
				if retrier.Attempt() > float64(do.cfg.maxRetries) {
					return nil, err
				}
				time.Sleep(retrier.Duration())
				continue
			}

			for _, log := range logs {
				ac, err := do.convertLogToCreditEvent(log)
				if err != nil {
					return nil, err
				}

				events = append(events, ac)
			}
			retrier.Reset()
			break
		}
	}
	return events, nil
}

func (do *EthDepositOracle) convertLogToCreditEvent(log types.Log) (*AccountCredit, error) {
	if (log.Topics == nil || len(log.Topics) == 0) || log.Topics[0] != do.creditEventSignature {
		do.logger.Debug("Log is not a credit event", zap.Any("log", log.Topics))
		return nil, fmt.Errorf("invalid credit event log")
	}

	event, err := do.eventABI.Unpack("Credit", log.Data)
	if err != nil {
		return nil, err
	}

	return &AccountCredit{
		Account:   event[0].(common.Address).Hex(),
		Amount:    event[1].(*big.Int),
		TxHash:    log.TxHash.String(),
		BlockHash: log.BlockHash.String(),
		ChainID:   do.cfg.chainID,
	}, nil
}

func (do *EthDepositOracle) addEvent(ctx context.Context, credit *AccountCredit) error {
	do.logger.Debug("Adding credit event to eventstore", zap.Any("event", credit))

	if credit == nil {
		return nil
	}

	bts, err := credit.MarshalBinary()
	if err != nil {
		return err
	}

	err = do.eventstore.Store(ctx, bts, credit.Type())
	if err != nil {
		return err
	}

	return nil
}

func (do *EthDepositOracle) getBlockHeight(ctx context.Context) (int64, error) {
	blockBytes, err := do.kvstore.Get(ctx, []byte(lastProcessedBlock))
	if err == kv.ErrKeyNotFound || blockBytes == nil {
		do.setBlockHeight(ctx, 0)
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	height := binary.BigEndian.Uint64(blockBytes)
	return int64(height), nil
}

func (do *EthDepositOracle) setBlockHeight(ctx context.Context, height int64) error {
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(height))
	return do.kvstore.Set(ctx, []byte(lastProcessedBlock), heightBytes)
}
