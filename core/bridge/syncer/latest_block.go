package syncer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/chain"
	"go.uber.org/zap"
)

// GetLatestBlock returns the latest block number that has enough confirmations.
func (b *blockSyncer) LatestBlock(ctx context.Context) (*chain.Header, error) {
	// this involes 2 calls; one to get the latest block and one to get the latest finalized block
	header, err := b.chainClient.HeaderByNumber(ctx, nil)
	if err != nil {
		b.log.Error("failed to get latest block", zap.Error(err))
		return nil, fmt.Errorf("failed to get latest block: %v", err)
	}

	lastFinalized := header.Height - b.requiredConfirmations
	if lastFinalized < 0 {
		b.log.Error("latest block is less than required confirmations", zap.Int64("latest block", header.Height),
			zap.Int64("required confirmations", b.requiredConfirmations),
			zap.Error(err))

		return nil, fmt.Errorf("latest block is less than required confirmations.  latest block: %d.  required confirmations: %d: %v", header.Height, b.requiredConfirmations, err)
	}

	bigLastFinalized := big.NewInt(lastFinalized)

	finalizedHeader, err := b.chainClient.HeaderByNumber(ctx, bigLastFinalized)
	if err != nil {
		b.log.Error("failed to get latest finalized block", zap.Error(err))
		return nil, err
	}

	return finalizedHeader, nil
}

func (b *blockSyncer) setLatestBlock(ctx context.Context) error {
	latest, err := b.LatestBlock(ctx)
	if err != nil {
		return err
	}

	b.lastBlock = latest.Height

	return nil
}
