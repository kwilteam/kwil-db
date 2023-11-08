package syncer

import (
	"time"
)

type BlockSyncerOpts func(*blockSyncer)

func WithReconnectInterval(intervalSeconds int64) BlockSyncerOpts {
	return func(b *blockSyncer) {
		b.reconnectInterval = time.Duration(intervalSeconds) * time.Second
	}
}

func WithRequiredConfirmations(confirmations int64) BlockSyncerOpts {
	return func(b *blockSyncer) {
		b.requiredConfirmations = confirmations
	}
}

func WithLastBlock(lastBlock int64) BlockSyncerOpts {
	return func(b *blockSyncer) {
		b.lastBlock = lastBlock
	}
}
