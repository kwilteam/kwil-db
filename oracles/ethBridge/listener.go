package ethbridge

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

func (eb *EthBridge) listen(ctx context.Context) error {
	// Listen for new blocks
	headers := make(chan *types.Header)
	sub, err := eb.ethclient.SubscribeNewHead(ctx, headers)
	if err != nil {
		return err
	}

	go func(ctx context.Context, sub ethereum.Subscription) {
		defer sub.Unsubscribe()
		lastHeight := eb.cfg.startingHeight
		requiredConfirmations := eb.cfg.requiredConfirmations
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				if err != nil {
					eb.logger.Warn("subscription error", zap.Error(err))
					sub, err = eb.resubscribe(ctx, sub)
					if err != nil {
						eb.logger.Error("Failed to resubscribe", zap.Error(err))
						return
					}

				}

			case header := <-headers:
				currentHeight := header.Number.Int64()
				eb.logger.Info("New block", zap.Int64("height", currentHeight))

				if currentHeight-lastHeight < requiredConfirmations {
					continue
				}

				// get all the deposit events
				FromBlock := big.NewInt(lastHeight + 1)
				ToBlock := big.NewInt(currentHeight - requiredConfirmations)
				events, err := eb.filterLogs(ctx, FromBlock, ToBlock)
				if err != nil {
					eb.logger.Error("Failed to filter logs", zap.Error(err))
					continue
				}

				for _, event := range events {
					eb.addEvent(ctx, event)
				}
				lastHeight = currentHeight - requiredConfirmations
			}
		}
	}(ctx, sub)

	return nil
}

func (eb *EthBridge) resubscribe(ctx context.Context, sub ethereum.Subscription) (ethereum.Subscription, error) {
	sub.Unsubscribe()

	headers := make(chan *types.Header)
	sub, err := eb.ethclient.SubscribeNewHead(ctx, headers)
	if err != nil {
		eb.logger.Error("Failed to resubscribe", zap.Error(err))
		return nil, err
	}

	return sub, nil
}

func (eb *EthBridge) filterLogs(ctx context.Context, from *big.Int, to *big.Int) ([]AccountCredit, error) {
	query := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   to,
		Addresses: []common.Address{common.HexToAddress(eb.cfg.escrowAddress)},
		Topics:    [][]common.Hash{{eb.creditEventSignature}},
	}
	logs, err := eb.ethclient.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	events := make([]AccountCredit, len(logs))
	for i, log := range logs {
		event, err := eb.eventABI.Unpack("Credit", log.Data)
		if err != nil {
			return nil, err
		}
		events[i] = AccountCredit{
			Account:   event[0].(common.Address).Hex(),
			Amount:    event[1].(*big.Int),
			TxHash:    log.TxHash.String(),
			BlockHash: log.BlockHash.String(),
			ChainID:   eb.cfg.chainID,
		}
	}

	return events, nil
}

func (eb *EthBridge) addEvent(ctx context.Context, credit AccountCredit) error {
	eb.logger.Debug("Adding credit event to eventstore", zap.Any("event", credit))

	bts, err := credit.MarshalBinary()
	if err != nil {
		return err
	}

	err = eb.eventstore.Store(ctx, bts, credit.Type())
	if err != nil {
		return err
	}

	return nil
}
