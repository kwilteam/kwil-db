package node

import (
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
)

// Config is the configuration for a [Node] instance.
type Config struct {
	RootDir string
	PrivKey crypto.PrivateKey
	Logger  log.Logger

	Genesis   config.GenesisConfig
	Consensus config.ConsensusConfig
	P2P       config.PeerConfig
	PG        config.PGConfig
}
