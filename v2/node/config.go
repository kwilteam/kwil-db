package node

import (
	"kwil/config"
	"kwil/crypto"
	"kwil/log"
	"kwil/node/types"
	ktypes "kwil/types"

	"github.com/libp2p/go-libp2p/core/host"
)

// Config is the configuration for a [Node] instance.
type Config struct {
	RootDir string
	PrivKey crypto.PrivateKey

	Cfg     *config.Config
	Genesis *config.GenesisConfig
	ValSet  map[string]ktypes.Validator

	Host       host.Host
	PeerMgr    PeerManager
	Mempool    types.MemPool
	BlockStore types.BlockStore
	Consensus  ConsensusEngine
	Logger     log.Logger
}
