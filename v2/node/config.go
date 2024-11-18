package node

import (
	"kwil/config"
	"kwil/crypto"
	"kwil/log"
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
