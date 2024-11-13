package node

import (
	"kwil/node/types"

	"github.com/libp2p/go-libp2p/core/host"
)

type options struct {
	// dependency overrides
	host host.Host
	bs   types.BlockStore
	mp   types.MemPool
	ce   ConsensusEngine
}

type Option func(*options)

func WithHost(host host.Host) Option {
	return func(o *options) {
		o.host = host
	}
}

func WithBlockStore(bs types.BlockStore) Option {
	return func(o *options) {
		o.bs = bs
	}
}

func WithMemPool(mp types.MemPool) Option {
	return func(o *options) {
		o.mp = mp
	}
}

func WithConsensusEngine(ce ConsensusEngine) Option {
	return func(o *options) {
		o.ce = ce
	}
}
