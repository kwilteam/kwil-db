package node

import (
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
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
