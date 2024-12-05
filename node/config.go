package node

import (
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/types"
)

// Config is the configuration for a [Node] instance.
type Config struct {
	RootDir string
	PrivKey crypto.PrivateKey
	DB      DB

	P2P       *config.PeerConfig
	DBConfig  *config.DBConfig
	Statesync *config.StateSyncConfig

	Mempool     types.MemPool
	BlockStore  types.BlockStore
	Consensus   ConsensusEngine
	Snapshotter SnapshotStore
	Logger      log.Logger
}
