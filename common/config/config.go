// Package config contains Kwil's config structures.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	merge "dario.cat/mergo"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
)

type KwildConfig struct {
	RootDir string

	AppConfig   *AppConfig   `mapstructure:"app"`
	ChainConfig *ChainConfig `mapstructure:"chain"`
	Logging     *Logging     `mapstructure:"log"`
}

type Logging struct {
	Level          string   `mapstructure:"level"`
	RPCLevel       string   `mapstructure:"rpc_level"`
	ConsensusLevel string   `mapstructure:"consensus_level"`
	DBLevel        string   `mapstructure:"db_level"`
	Format         string   `mapstructure:"format"`
	TimeEncoding   string   `mapstructure:"time_format"`
	OutputPaths    []string `mapstructure:"output_paths"`
	MaxLogSizeKB   int64    `mapstructure:"file_roll_size"`
	MaxLogRolls    int      `mapstructure:"retain_max_rolls"`
}

type AppConfig struct {
	JSONRPCListenAddress string `mapstructure:"jsonrpc_listen_addr"`
	AdminListenAddress   string `mapstructure:"admin_listen_addr"`

	PrivateKeyPath string `mapstructure:"private_key_path"`

	// PostgreSQL DB settings. DBName is the name if the PostgreSQL database to
	// connect to. The different data stores (e.g. engine, acct store, event
	// store, etc.) are all in the same database. Assuming "kwild" is the
	// DBName, this would be created with psql with the commands:
	//  CREATE USER kwild WITH SUPERUSER REPLICATION;
	//  CREATE DATABASE kwild OWNER kwild;
	//
	// All of these settings are strings and separate, but it is possible to
	// have a single DB "connection string" to pass to the PostgreSQL backend.
	// However, this is less error prone, and prevents passing settings that
	// would alter the functionality of the connection. An advanced option could
	// be added to supplement the conn string if that seems useful.
	DBHost string `mapstructure:"pg_db_host"`
	DBPort string `mapstructure:"pg_db_port"`
	DBUser string `mapstructure:"pg_db_user"`
	DBPass string `mapstructure:"pg_db_pass"`
	DBName string `mapstructure:"pg_db_name"`

	RPCTimeout         Duration                     `mapstructure:"rpc_timeout"`
	RPCMaxReqSize      int                          `mapstructure:"rpc_req_limit"`
	PrivateRPC         bool                         `mapstructure:"private_rpc"`
	ChallengeExpiry    Duration                     `mapstructure:"challenge_expiry"`
	ChallengeRateLimit float64                      `mapstructure:"challenge_rate_limit"`
	ReadTxTimeout      Duration                     `mapstructure:"db_read_timeout"`
	ExtensionEndpoints []string                     `mapstructure:"extension_endpoints"`
	AdminRPCPass       string                       `mapstructure:"admin_pass"`
	NoTLS              bool                         `mapstructure:"admin_notls"`
	TLSCertFile        string                       `mapstructure:"tls_cert_file"`
	TLSKeyFile         string                       `mapstructure:"tls_key_file"`
	Hostname           string                       `mapstructure:"hostname"`
	ProfileMode        string                       `mapstructure:"profile_mode"`
	ProfileFile        string                       `mapstructure:"profile_file"`
	Extensions         map[string]map[string]string `mapstructure:"extensions"`

	Snapshots SnapshotConfig `mapstructure:"snapshots"`

	// GenesisState is the path to the snapshot file containing genesis state
	// to be loaded on startup during network initialization. If genesis app_hash
	// is not provided, this snapshot file is not used.
	GenesisState string `mapstructure:"genesis_state"`

	// MigrateFrom is the JSON-RPC listening address of the node to replicate the state from.
	MigrateFrom string `mapstructure:"migrate_from"`
}

type SnapshotConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	RecurringHeight uint64 `mapstructure:"recurring_height"`
	MaxSnapshots    uint64 `mapstructure:"max_snapshots"`
	SnapshotDir     string `mapstructure:"snapshot_dir"`
}

type ChainRPCConfig struct {
	// TCP or UNIX socket address for the RPC server to listen on
	ListenAddress string `mapstructure:"listen_addr"`

	// How long to wait for a tx to be committed when transactions are authored with --sync flag
	BroadcastTxTimeout Duration `mapstructure:"broadcast_tx_timeout"`
}

type P2PConfig struct {
	// ListenAddress is the address on which to listen for incoming connections.
	ListenAddress string `mapstructure:"listen_addr"`
	// ExternalAddress is the address to advertise to peers to dial us.
	ExternalAddress string `mapstructure:"external_address"`
	// PersistentPeers is a comma separated list of nodes to keep persistent
	// connections to.
	PersistentPeers string `mapstructure:"persistent_peers"`
	// PrivateMode prevents other nodes from connecting to the node unless
	// they are the current validators or a part of the whitelistPeers.
	// If disabled, the node by default operates in public mode, where any node can connect to it.
	PrivateMode bool `mapstructure:"private_mode"`
	// WhitelistPeers is a comma separated list of nodeIDs that can connect to this node.
	// This is excluding any persistent peers or seeds or current validators.
	WhitelistPeers string `mapstructure:"whitelist_peers"`
	// AddrBookStrict enforces strict address routability rules. This must be
	// false for private or local networks.
	AddrBookStrict bool `mapstructure:"addr_book_strict"`
	// MaxNumInboundPeers is the maximum number of inbound peers.
	MaxNumInboundPeers int `mapstructure:"max_num_inbound_peers"`
	// MaxNumOutboundPeers is the maximum number of outbound peers to connect
	// to, excluding persistent peers.
	MaxNumOutboundPeers int `mapstructure:"max_num_outbound_peers"`
	// UnconditionalPeerIDs are the node IDs to which a connection will be
	// (re)established ignoring any existing limits.
	UnconditionalPeerIDs string `mapstructure:"unconditional_peer_ids"`
	// PexReactor enables the peer-exchange reactor.
	PexReactor bool `mapstructure:"pex"`
	// AllowDuplicateIP permits peers connecting from the same IP.
	AllowDuplicateIP bool `mapstructure:"allow_duplicate_ip"`
	// HandshakeTimeout is the peer connection handshake timeout.
	HandshakeTimeout Duration `mapstructure:"handshake_timeout"`
	// DialTimeout is the peer connection establishment timeout.
	DialTimeout Duration `mapstructure:"dial_timeout"`
	// SeedMode makes the node constantly crawls the network looking for peers.
	// If another node asks it for addresses, it responds and disconnects.
	// Requires peer-exchange.
	SeedMode bool `mapstructure:"seed_mode"`
	// Seeds is a comma-separated separated list of seed nodes to query for peer
	// addresses. Only used if the peers in the address book are unreachable.
	Seeds string `mapstructure:"seeds"`
}

type MempoolConfig struct {
	// Maximum number of transactions in the mempool
	Size int `mapstructure:"size"`
	// Size of the cache (used to filter transactions we saw earlier) in transactions
	CacheSize int `mapstructure:"cache_size"`

	// MaxTxBytes limits the size of any one transaction in mempool.
	MaxTxBytes int `mapstructure:"max_tx_bytes"`

	// MaxTxsBytes limits the total size of all txs in the mempool.
	// This only accounts for raw transactions (e.g. given 1MB transactions and
	// max_txs_bytes=5MB, mempool will only accept 5 transactions).
	MaxTxsBytes int `mapstructure:"max_txs_bytes"`
}

type ConsensusConfig struct {
	// TimeoutPropose is how long to wait for a proposal block before prevoting
	// nil.
	TimeoutPropose Duration `mapstructure:"timeout_propose"`
	// TimeoutPrevote is how long to wait after receiving +2/3 prevotes for
	// “anything” (i.e. not a single block or nil).
	TimeoutPrevote Duration `mapstructure:"timeout_prevote"`
	// TimeoutPrecommit is how long we wait after receiving +2/3 precommits for
	// “anything” (i.e. not a single block or nil).
	TimeoutPrecommit Duration `mapstructure:"timeout_precommit"`
	// TimeoutCommit is how long to wait after committing a block, before
	// starting on the new height (this gives us a chance to receive some more
	// precommits, even though we already have +2/3).
	TimeoutCommit Duration `mapstructure:"timeout_commit"`
}

type StateSyncConfig struct {
	Enable bool `mapstructure:"enable"`

	// SnapshotDir is the directory to store the received snapshot chunks.
	SnapshotDir string `mapstructure:"snapshot_dir"`

	// Trusted snapshot servers to fetch/validate the snapshots from.
	// At least 1 server is required for the state sync to work.
	RPCServers string `mapstructure:"rpc_servers"`

	// Time to spend discovering snapshots before initiating starting
	// the db restoration using snapshot.
	DiscoveryTime Duration `mapstructure:"discovery_time"`

	// The timeout duration before re-requesting a chunk, possibly from a different
	// peer (default: 1 minute).
	ChunkRequestTimeout Duration `mapstructure:"chunk_request_timeout"`

	// Light client verification options, Automatically fetched from the RPC Servers
	// during the node initialization.
	TrustHeight int64
	TrustHash   string
	TrustPeriod Duration
}

type ChainConfig struct {
	Moniker string `mapstructure:"moniker"`
	// DBPath  string `mapstructure:"db_dir"` // internal/abci knows this

	RPC       *ChainRPCConfig  `mapstructure:"rpc"`
	P2P       *P2PConfig       `mapstructure:"p2p"`
	Mempool   *MempoolConfig   `mapstructure:"mempool"`
	StateSync *StateSyncConfig `mapstructure:"statesync"`
	Consensus *ConsensusConfig `mapstructure:"consensus"`
}

// toml package does not support time.Duration, since time is not part of TOML spec
// Fix can be found here: https://github.com/pelletier/go-toml/issues/767
// It implements both the TextUnmarshaler interface and the pflag.Value interface
type Duration time.Duration

var _ pflag.Value = (*Duration)(nil)

func (d Duration) Dur() time.Duration {
	return time.Duration(d)
}

func (d *Duration) UnmarshalText(b []byte) error {
	x, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

func (d *Duration) String() string {
	// if not set, we need to return an empty string,
	// so that the -h flag does not show it as a default
	// value of 0s
	if d == nil {
		return ""
	}
	if *d == 0 {
		return ""
	}
	return time.Duration(*d).String()
}

func (d *Duration) Type() string {
	return "duration"
}

func (d *Duration) Set(s string) error {
	x, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

// Merge merges b onto a, overwriting any fields in a that are also set in b.
func (a *KwildConfig) Merge(b *KwildConfig) error {
	return merge.MergeWithOverwrite(a, b)
}

func (a *KwildConfig) MarshalBinary() ([]byte, error) {
	mapCfg := make(map[string]interface{})
	mapstructure.Decode(a, &mapCfg)
	return json.Marshal(mapCfg)
}

func (a *KwildConfig) UnmarshalBinary(b []byte) error {
	mapCfg := make(map[string]interface{})
	err := json.Unmarshal(b, &mapCfg)
	if err != nil {
		return err
	}
	return mapstructure.Decode(mapCfg, a)
}

func (cfg *KwildConfig) LogConfig() (*log.Config, error) {
	// Rootify any relative paths.
	outputPaths := make([]string, 0, len(cfg.Logging.OutputPaths))
	for _, path := range cfg.Logging.OutputPaths {
		switch path {
		case "stdout", "stderr":
			outputPaths = append(outputPaths, path)
		default:
			updatedPath, err := CleanPath(path, cfg.RootDir)
			if err != nil {
				return nil, err
			}
			outputPaths = append(outputPaths, updatedPath)
		}
	}
	// log.Config <== Logging
	return &log.Config{
		Level:        cfg.Logging.Level,
		OutputPaths:  outputPaths,
		Format:       cfg.Logging.Format,
		EncodeTime:   cfg.Logging.TimeEncoding,
		MaxLogSizeKB: cfg.Logging.MaxLogSizeKB,
		MaxLogRolls:  cfg.Logging.MaxLogRolls,
	}, nil
}

// configureExtensions sets up the extensions for the node.
func (cfg *KwildConfig) ConfigureExtensions(genCfg *chain.GenesisConfig) error {
	extensions := cfg.AppConfig.Extensions

	// Migrations extension configuration
	// sets the listener address from the migrate_from flag
	startHeight := genCfg.ConsensusParams.Migration.StartHeight
	endHeight := genCfg.ConsensusParams.Migration.EndHeight
	// validate migration configuration
	// migration start and end heights must be set together
	if (startHeight > 0 && endHeight == 0) || (startHeight == 0 && endHeight > 0) {
		return fmt.Errorf("migration configuration requires both start and end heights to be set")
	}

	// if migration start and end heights are set, MigrateFrom must be set
	if (startHeight > 0 && endHeight > 0) && cfg.AppConfig.MigrateFrom == "" {
		return fmt.Errorf("migrate_from flag must be set when migration configuration is set in the genesis file")
	}

	// and vice versa
	if cfg.AppConfig.MigrateFrom != "" && (startHeight == 0 || endHeight == 0) {
		return fmt.Errorf("migrate_from flag requires migration start and end heights in the genesis file")
	}

	if cfg.AppConfig.MigrateFrom != "" {
		extensions["migrations"] = map[string]string{
			"listen_address": cfg.AppConfig.MigrateFrom,
			"start_height":   fmt.Sprintf("%d", startHeight),
			"end_height":     fmt.Sprintf("%d", endHeight),
		}
	}

	// other extensions can be configured here
	cfg.AppConfig.Extensions = extensions
	return nil
}

// CleanPath returns an absolute path for the given path, relative to the root directory.
// It detects paths starting with ~/ and expands them to the user's home directory.
func CleanPath(path, rootDir string) (string, error) {
	// If the path is already absolute, return it as is.
	if filepath.IsAbs(path) {
		return path, nil
	}

	// If the path is ~/..., expand it to the user's home directory.
	if tail, cut := strings.CutPrefix(path, "~/"); cut {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, tail), nil
	}

	// Otherwise, treat it as relative to the root directory.
	return filepath.Join(rootDir, path), nil
}
