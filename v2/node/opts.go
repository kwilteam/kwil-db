package node

import (
	"kwil/log"
	"kwil/node/types"

	"github.com/libp2p/go-libp2p/core/host"
)

type options struct {
	ip      string // netip.Addr maybe
	port    uint64
	privKey []byte
	role    types.Role
	pex     bool
	logger  log.Logger
	host    host.Host
	bs      types.BlockStore
	mp      types.MemPool
	ce      ConsensusEngine
	valSet  map[string]types.Validator
	leader  []byte
}

type Option func(*options)

func WithLogger(logger log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithIP(ip string) Option {
	return func(o *options) {
		o.ip = ip
	}
}

func WithPort(port uint64) Option {
	return func(o *options) {
		o.port = port
	}
}

func WithPrivKey(privKey []byte) Option {
	return func(o *options) {
		o.privKey = privKey
	}
}

func WithRole(role types.Role) Option {
	return func(o *options) {
		o.role = role
	}
}

func WithPex(pex bool) Option {
	return func(o *options) {
		o.pex = pex
	}
}

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

func WithGenesisValidators(valSet map[string]types.Validator) Option {
	return func(o *options) {
		o.valSet = valSet
	}
}

func WithLeader(leader []byte) Option {
	return func(o *options) {
		o.leader = leader
	}
}
