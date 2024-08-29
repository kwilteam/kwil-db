package cmd

import (
	"time"

	commonConfig "github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/core/log"
)

// binaryConfig configures the generated binary. It is able to control the binary names.
// It is primarily used for generating useful help commands that have proper names.
type binaryConfig struct {
	// RootCmd is the name of the root command.
	// If we are building kwild / kwil-cli / kwil-admin, then
	// RootCmd is empty.
	RootCmd string
	// NodeCmd is the name of the node command.
	NodeCmd string
	// ClientCmd is the name of the client command.
	ClientCmd string
	// AdminCmd is the name of the admin command.
	AdminCmd string
	// ProjectName is the name of the project, which will be used in the help text.
	ProjectName string
}

var BinaryConfig = defaultBinaryConfig()

func (b *binaryConfig) NodeUsage() string {
	if b.RootCmd != "" {
		return b.RootCmd + " " + b.NodeCmd
	}
	return b.NodeCmd
}

func (b *binaryConfig) ClientUsage() string {
	if b.RootCmd != "" {
		return b.RootCmd + " " + b.ClientCmd
	}
	return b.ClientCmd
}

func (b *binaryConfig) AdminUsage() string {
	if b.RootCmd != "" {
		return b.RootCmd + " " + b.AdminCmd
	}
	return b.AdminCmd
}

func defaultBinaryConfig() binaryConfig {
	return binaryConfig{
		ProjectName: "Kwil",
		NodeCmd:     "kwild",
		ClientCmd:   "kwil-cli",
		AdminCmd:    "kwil-admin",
	}
}

// DefaultConfig returns the default configuration for kwild.
// It is exported as a function so that users can customize the default configuration.
var DefaultConfig = func() *commonConfig.KwildConfig {
	return &commonConfig.KwildConfig{
		AppConfig: &commonConfig.AppConfig{
			JSONRPCListenAddress: "0.0.0.0:8484",
			AdminListenAddress:   "/tmp/kwild.socket", // Or, suggested, 127.0.0.1:8485
			PrivateKeyPath:       "private_key",
			DBHost:               "127.0.0.1",
			DBPort:               "5432", // ignored with unix socket, but applies if IP used for DBHost
			DBUser:               "kwild",
			DBName:               "kwild",
			RPCTimeout:           commonConfig.Duration(45 * time.Second),
			RPCMaxReqSize:        4_200_000,
			ChallengeExpiry:      commonConfig.Duration(10 * time.Second),
			ChallengeRateLimit:   10.0, // req/s
			ReadTxTimeout:        commonConfig.Duration(5 * time.Second),
			Extensions:           make(map[string]map[string]string),
			Snapshots: commonConfig.SnapshotConfig{
				Enabled:         false,
				RecurringHeight: 14400, // 1 day at 6s block time
				MaxSnapshots:    3,
				SnapshotDir:     "snapshots",
				MaxRowSize:      4 * 1024 * 1024,
			},
			GenesisState: "",
		},
		Logging: &commonConfig.Logging{
			Level:        "info",
			Format:       log.FormatJSON,
			TimeEncoding: log.TimeEncodingEpochFloat,
			OutputPaths:  []string{"stdout", "kwild.log"},
		},

		ChainConfig: &commonConfig.ChainConfig{
			P2P: &commonConfig.P2PConfig{
				ListenAddress:       "tcp://0.0.0.0:26656",
				ExternalAddress:     "",
				PrivateMode:         false,
				AddrBookStrict:      false, // override comet
				MaxNumInboundPeers:  40,
				MaxNumOutboundPeers: 10,
				AllowDuplicateIP:    true, // override comet
				PexReactor:          true,
				HandshakeTimeout:    commonConfig.Duration(20 * time.Second),
				DialTimeout:         commonConfig.Duration(3 * time.Second),
			},
			RPC: &commonConfig.ChainRPCConfig{
				ListenAddress:      "tcp://127.0.0.1:26657",
				BroadcastTxTimeout: commonConfig.Duration(15 * time.Second), // 2.5x default TimeoutCommit (6s)
			},
			Mempool: &commonConfig.MempoolConfig{
				Size:        50000,
				CacheSize:   60000,
				MaxTxBytes:  1024 * 1024 * 4,   // 4 MiB
				MaxTxsBytes: 1024 * 1024 * 512, // 512 MiB
			},
			StateSync: &commonConfig.StateSyncConfig{
				Enable:              false,
				SnapshotDir:         "rcvdSnaps",
				DiscoveryTime:       commonConfig.Duration(15 * time.Second),
				ChunkRequestTimeout: commonConfig.Duration(10 * time.Second),
				TrustPeriod:         commonConfig.Duration(36000 * time.Second),
			},
			Consensus: &commonConfig.ConsensusConfig{
				TimeoutPropose:   commonConfig.Duration(3 * time.Second),
				TimeoutPrevote:   commonConfig.Duration(2 * time.Second),
				TimeoutPrecommit: commonConfig.Duration(2 * time.Second),
				TimeoutCommit:    commonConfig.Duration(6 * time.Second),
			},
		},
	}
}
