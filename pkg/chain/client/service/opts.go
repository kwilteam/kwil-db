package service

import (
	"github.com/kwilteam/kwil-db/pkg/chain/types"
	"github.com/kwilteam/kwil-db/pkg/log"
	"time"
)

type ChainClientOpts func(*chainClient)

func WithReconnectInterval(intervalSeconds int64) ChainClientOpts {
	return func(c *chainClient) {
		c.reconnectInterval = time.Duration(intervalSeconds) * time.Second
	}
}

func WithRequiredConfirmations(confirmations int64) ChainClientOpts {
	return func(c *chainClient) {
		c.requiredConfirmations = confirmations
	}
}

func WithChainCode(chainCode types.ChainCode) ChainClientOpts {
	return func(c *chainClient) {
		c.chainCode = types.ChainCode(chainCode)
	}
}

func WithLastBlock(lastBlock int64) ChainClientOpts {
	return func(c *chainClient) {
		c.lastBlock = lastBlock
	}
}

func WithLogger(logger log.Logger) ChainClientOpts {
	return func(c *chainClient) {
		c.log = logger
	}
}
