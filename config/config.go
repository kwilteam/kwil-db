package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"

	"github.com/pelletier/go-toml/v2"
)

const (
	ConfigFileName  = "kwil.toml"
	GenesisFileName = "genesis.json"
)

// Duration is a wrapper around time.Duration that implements text
// (un)marshalling for the go-toml package to work with Go duration strings
// instead of integers.
type Duration time.Duration

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalText(text []byte) error {
	duration, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

type GenesisConfig struct {
	ChainID       string `json:"chain_id"`
	InitialHeight int64  `json:"initial_height"`
	// Leader is the leader's public key.
	Leader types.HexBytes `json:"leader"`
	// Validators is the list of genesis validators (including the leader).
	Validators []*types.Validator `json:"validators"`

	// MaxBlockSize is the maximum size of a block in bytes.
	MaxBlockSize int64 `json:"max_block_size"`
	// JoinExpiry is the number of blocks after which the validators
	// join request expires if not approved.
	JoinExpiry int64 `json:"join_expiry"`
	// VoteExpiry is the default number of blocks after which the validators
	// vote expires if not approved.
	VoteExpiry int64 `json:"vote_expiry"`
	// DisabledGasCosts dictates whether gas costs are disabled.
	DisabledGasCosts bool `json:"disabled_gas_costs"`
	// MaxVotesPerTx is the maximum number of votes that can be included in a
	// single transaction.
	MaxVotesPerTx int64 `json:"max_votes_per_tx"`
	// StateHash is the hash of the initial state of the chain, used when bootstrapping
	// the chain with a network snapshot.
	StateHash []byte `json:"state_hash"`
}

func (nc *GenesisConfig) SaveAs(filename string) error {
	bts, err := json.MarshalIndent(nc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, bts, 0644)
}

func LoadGenesisConfig(filename string) (*GenesisConfig, error) {
	bts, err := os.ReadFile(filename)
	if err != nil {
		return nil, err // can be os.ErrNotExist
	}

	var nc GenesisConfig
	if err := json.Unmarshal(bts, &nc); err != nil {
		return nil, err
	}

	return &nc, nil
}

func DefaultGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		ChainID:          "kwil-test-chain",
		InitialHeight:    0,
		Leader:           types.HexBytes{},
		Validators:       []*types.Validator{},
		DisabledGasCosts: true,
		JoinExpiry:       14400,
		VoteExpiry:       108000,
		MaxBlockSize:     6 * 1024 * 1024,
		MaxVotesPerTx:    200,
	}
}

// const (
// 	defaultUserRPCPort  = 8484
// 	defaultAdminRPCPort = 8584
// 	defaultP2PRPCPort   = 6600
// )

// DefaultConfig generates an instance of the default config.
func DefaultConfig() *Config {
	return &Config{
		LogLevel:  log.LevelInfo,
		LogFormat: log.FormatUnstructured,
		// Private key is empty by default. This should probably be moved back
		// to a key file again to avoid accidentally leaking the key.
		P2P: PeerConfig{
			IP:        "0.0.0.0",
			Port:      6600,
			Pex:       true,
			BootNodes: []string{},
		},
		Consensus: ConsensusConfig{
			ProposeTimeout: Duration(1000 * time.Millisecond),
			MaxBlockSize:   50_000_000,
			MaxTxsPerBlock: 20_000,
		},
		DB: DBConfig{
			Host:          "127.0.0.1",
			Port:          "5432",
			User:          "kwild",
			Pass:          "",
			DBName:        "kwild",
			ReadTxTimeout: Duration(45 * time.Second),
			MaxConns:      60,
		},
		RPC: RPCConfig{
			ListenAddress:      "0.0.0.0:8484",
			Timeout:            Duration(20 * time.Second),
			MaxReqSize:         6_000_000,
			Private:            false,
			ChallengeExpiry:    Duration(30 * time.Second),
			ChallengeRateLimit: 10,
		},
		Admin: AdminConfig{
			Enable:        true,
			ListenAddress: "/tmp/kwild.socket",
			Pass:          "",
			NoTLS:         false,
			TLSCertFile:   "admin.cert",
			TLSKeyFile:    "admin.key",
		},
		Snapshots: SnapshotConfig{
			Enable:          false,
			RecurringHeight: 14400,
			MaxSnapshots:    3,
		},
		StateSync: StateSyncConfig{
			Enable:           false,
			DiscoveryTimeout: Duration(30 * time.Second),
			MaxRetries:       3,
		},
	}
}

// Config is the node's config.
type Config struct {
	LogLevel  log.Level  `toml:"log_level" comment:"log level\npossible values: 'debug', 'info', 'warn', and 'error'"`
	LogFormat log.Format `toml:"log_format" comment:"log format\npossible values: 'json', 'text' (kv), and 'plain' (fmt-style)"`
	// LogOutput []string   `toml:"log_output" comment:"output paths for the log"`

	PrivateKey types.HexBytes `toml:"privkey" comment:"private key to use for node"`

	// ProfileMode string `toml:"profile_mode"`
	// ProfileFile string `toml:"profile_file"`

	P2P       PeerConfig      `toml:"p2p" comment:"P2P related configuration"`
	Consensus ConsensusConfig `toml:"consensus" comment:"Consensus related configuration"`
	DB        DBConfig        `toml:"db" comment:"DB (PostgreSQL) related configuration"`
	RPC       RPCConfig       `toml:"rpc" comment:"User RPC service configuration"`
	Admin     AdminConfig     `toml:"admin" comment:"Admin RPC service configuration"`
	Snapshots SnapshotConfig  `toml:"snapshots" comment:"Snapshot creation and provider configuration"`
	StateSync StateSyncConfig `toml:"state_sync" comment:"Statesync configuration (vs block sync)"`
}

// PeerConfig corresponds to the [peer] section of the config.
type PeerConfig struct {
	IP        string   `toml:"ip" comment:"IP address to listen on for P2P connections"`
	Port      uint64   `toml:"port" comment:"port to listen on for P2P connections"`
	Pex       bool     `toml:"pex" comment:"enable peer exchange"`
	BootNodes []string `toml:"bootnodes" comment:"bootnodes to connect to on startup"`
}

type DBConfig struct {
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
	Host          string   `toml:"host" comment:"postgres host name (IP or UNIX socket path)"`
	Port          string   `toml:"port" comment:"postgres TCP port (leave empty for UNIX socket)"`
	User          string   `toml:"user" comment:"postgres role/user name"`
	Pass          string   `toml:"pass" comment:"postgres password if required for the user and host"`
	DBName        string   `toml:"dbname" comment:"postgres database name"`
	ReadTxTimeout Duration `toml:"read_timeout" comment:"timeout on read transactions from user RPC calls and queries"`
	MaxConns      uint32   `toml:"max_connections" comment:"maximum number of DB connections to permit"`
}

type ConsensusConfig struct {
	ProposeTimeout Duration `toml:"propose_timeout" comment:"timeout for proposing a block (applies to leader)"`
	MaxBlockSize   uint64   `toml:"max_block_size" comment:"max size of a block in bytes"`
	MaxTxsPerBlock uint64   `toml:"max_txs_per_block" comment:"max number of transactions per block"`
	// ? reannounce intervals?
}

type RPCConfig struct {
	ListenAddress      string   `toml:"listen" comment:"address in host:port format on which the RPC server will listen"`
	Timeout            Duration `toml:"timeout" comment:"user request duration limit after which it is cancelled"`
	MaxReqSize         int      `toml:"max_req_size" comment:"largest permissible user request size"`
	Private            bool     `toml:"private" comment:"enable private mode that requires challenge authentication for each call"`
	ChallengeExpiry    Duration `toml:"challenge_expiry" comment:"lifetime of a server-generated challenge"`
	ChallengeRateLimit float64  `toml:"challenge_rate_limit" comment:"maximum number of challenges per second that a user can request"`
}

type AdminConfig struct {
	Enable        bool   `toml:"enable" comment:"enable the admin RPC service"`
	ListenAddress string `toml:"listen" comment:"address in host:port format or UNIX socket path on which the admin RPC server will listen"`
	Pass          string `toml:"pass" comment:"optional password for the admin service"`
	NoTLS         bool   `toml:"notls" comment:"disable TLS when the listen address is not a loopback IP or UNIX socket"`
	TLSCertFile   string `toml:"cert" comment:"TLS certificate for use with a non-loopback listen address when notls is not true"`
	TLSKeyFile    string `toml:"key" comment:"TLS key for use with a non-loopback listen address when notls is not true"`
}

type SnapshotConfig struct {
	Enable          bool   `toml:"enable" comment:"enable creating and providing snapshots for peers using statesync"`
	RecurringHeight uint64 `toml:"recurring_height" comment:"snapshot creation period in blocks"`
	MaxSnapshots    uint64 `toml:"max_snapshots" comment:"number of snapshots to keep, after the oldest is removed when creating a new one"`
}

type StateSyncConfig struct {
	Enable           bool     `toml:"enable" comment:"enable using statesync rather than blocksync"`
	TrustedProviders []string `toml:"trusted_providers" comment:"trusted snapshot providers in node ID format (see bootnodes)"`

	DiscoveryTimeout Duration `toml:"discovery_time" comment:"how long to discover snapshots before selecting one to use"`
	MaxRetries       uint64   `toml:"max_retries" comment:"how many times to try after failing to apply a snapshot before switching to blocksync"`
}

// ConfigToTOML marshals the config to TOML. The `toml` struct field tag
// specifies the field names. For example:
//
//	Enable  bool  `toml:"enable,commented" comment:"enable the thing"`
//
// The above field will be written like:
//
//	# enable the thing
//	#enable=false
func (nc Config) ToTOML() ([]byte, error) {
	return toml.Marshal(nc)
}

func (nc *Config) FromTOML(b []byte) error {
	return toml.Unmarshal(b, &nc)
}

// SaveAs writes the Config to the specified TOML file.
func (nc *Config) SaveAs(filename string) error {
	bts, err := nc.ToTOML()
	if err != nil {
		return err
	}

	// TODO: write a toml header/comment or do some prettification. The template
	// was a maintenance burden, and we get setting and section comment with
	// field tags, so I do not prefer ea template. If it does not look pretty
	// enough, we may consider some basic post-processing of bts before writing
	// it. For example, insert newlines before each "#", write a header at the
	// top, etc.

	return os.WriteFile(filename, bts, 0644)
}

func LoadConfig(filename string) (*Config, error) {
	bts, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var nc Config
	if err := toml.Unmarshal(bts, &nc); err != nil {
		return nil, err
	}

	return &nc, nil
}
