package syncer

import (
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	bClient "github.com/kwilteam/kwil-db/oracles/tokenbridge/client"
)

const (
	// DefaultReconnectInterval is the default interval between reconnect attempts
	DefaultReconnectInterval = 30 * time.Second

	// DefaultRequiredConfirmations is the default number of confirmations required for a transaction to be considered final
	DefaultRequiredConfirmations = 12

	// DefaultLastBlock is the default last block.
	DefaultLastBlock = int64(0)
)

type blockSyncer struct {
	chainClient           ChainClient
	log                   log.Logger
	reconnectInterval     time.Duration
	requiredConfirmations int64
	lastBlock             int64
}

func New(chainClient bClient.ChainClient, opts ...BlockSyncerOpts) (*blockSyncer, error) {
	bs := &blockSyncer{
		log:                   log.NewNoOp(),
		reconnectInterval:     DefaultReconnectInterval,
		requiredConfirmations: DefaultRequiredConfirmations,
		lastBlock:             DefaultLastBlock,
	}

	for _, opt := range opts {
		opt(bs)
	}

	bs.chainClient = chainClient
	return bs, nil
}

func (b *blockSyncer) Close() error {
	return b.chainClient.Close()
}
