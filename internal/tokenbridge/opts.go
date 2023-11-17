package tokenbridge

import (
	"github.com/kwilteam/kwil-db/core/log"
)

type TokenBridgeOpts func(*TokenBridge)

func WithStartingHeight(height int64) TokenBridgeOpts {
	return func(b *TokenBridge) {
		b.startingHeight = height
	}
}

func WithLogger(logger log.Logger) TokenBridgeOpts {
	return func(b *TokenBridge) {
		b.log = logger
	}
}

func WithChunkSize(chunkSize int64) TokenBridgeOpts {
	return func(b *TokenBridge) {
		b.chunkSize = chunkSize
	}
}

func WithNodeAddress(nodeAddress string) TokenBridgeOpts {
	return func(b *TokenBridge) {
		b.nodeAddress = nodeAddress
	}
}
