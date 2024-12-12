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
		// Private key is empty by default.
		P2P: PeerConfig{
			IP:        "0.0.0.0",
			Port:      6600,
			Pex:       true,
			BootNodes: []string{},
		},
		Consensus: ConsensusConfig{
			ProposeTimeout: 1000 * time.Millisecond,
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
			Timeout:            20 * time.Second,
			MaxReqSize:         6_000_000,
			Private:            false,
			ChallengeExpiry:    30 * time.Second,
			ChallengeRateLimit: 10,
		},
		Admin: AdminConfig{
			Enable:        true,
			ListenAddress: "/tmp/kwil2-admin.socket",
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
			DiscoveryTimeout: 30 * time.Second,
			MaxRetries:       3,
		},
	}
}

// Config is the node's config.
type Config struct {
	// NOTE about tags:
	//
	//  - toml tags are used to marshal into a toml file with pelletier's go-toml
	//    (gotoml.Marshal: Config{} => []byte(tomlString))
	//
	//  - koanf tags are used to unmarshal into this struct from a koanf instance
	//    (k.Unmarshal: map[string]interface{} => Config{})
	//
	// Presently these tags are the same. If we change the canonicalization,
	// such as removing both dashes and underscores, the tags would be different.

	LogLevel  log.Level  `koanf:"log_level" toml:"log_level" comment:"log level"`
	LogFormat log.Format `koanf:"log_format" toml:"log_format" comment:"log format"`
	// LogOutput []string   `koanf:"log_output" toml:"log_output" comment:"output paths for the log"`

	PrivateKey types.HexBytes `koanf:"privkey" toml:"privkey" comment:"private key to use for node"`

	// ProfileMode string `koanf:"profile_mode" toml:"profile_mode"`
	// ProfileFile string `koanf:"profile_file" toml:"profile_file"`

	// subsections below

	P2P PeerConfig `koanf:"p2p" toml:"p2p"`

	Consensus ConsensusConfig `koanf:"consensus" toml:"consensus"`
	DB        DBConfig        `koanf:"db" toml:"db"`
	RPC       RPCConfig       `koanf:"rpc" toml:"rpc"`
	Admin     AdminConfig     `koanf:"admin" toml:"admin"`
	Snapshots SnapshotConfig  `koanf:"snapshots" toml:"snapshots"`
	StateSync StateSyncConfig `koanf:"state_sync" toml:"state_sync"`
}

// PeerConfig corresponds to the [peer] section of the config.
type PeerConfig struct {
	IP        string   `koanf:"ip" toml:"ip" comment:"ip to listen on for P2P connections"`
	Port      uint64   `koanf:"port" toml:"port" comment:"port to listen on for P2P connections"`
	Pex       bool     `koanf:"pex" toml:"pex" comment:"enable peer exchange"`
	BootNodes []string `koanf:"bootnodes" toml:"bootnodes" comment:"bootnodes to connect to on startup"`

	// ListenAddr string // "127.0.0.1:6600"
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
	Host          string   `koanf:"host" toml:"host"`
	Port          string   `koanf:"port" toml:"port"`
	User          string   `koanf:"user" toml:"user"`
	Pass          string   `koanf:"pass" toml:"pass"`
	DBName        string   `koanf:"dbname" toml:"dbname"`
	ReadTxTimeout Duration `koanf:"read_timeout" toml:"read_timeout"`
	MaxConns      uint32   `koanf:"max_connections" toml:"max_connections"`
}

type ConsensusConfig struct {
	ProposeTimeout time.Duration `koanf:"propose_timeout" toml:"propose_timeout" comment:"timeout for proposing a block"`
	MaxBlockSize   uint64        `koanf:"max_block_size" toml:"max_block_size" comment:"max size of a block in bytes"`
	MaxTxsPerBlock uint64        `koanf:"max_txs_per_block" toml:"max_txs_per_block" comment:"max number of transactions per block"`
	// ? reannounce intervals?
}

type RPCConfig struct {
	ListenAddress      string        `koanf:"listen" toml:"listen"`
	Timeout            time.Duration `koanf:"timeout" toml:"timeout"`
	MaxReqSize         int           `koanf:"max_req_size" toml:"max_req_size"`
	Private            bool          `koanf:"private" toml:"private"`
	ChallengeExpiry    time.Duration `koanf:"challenge_expiry" toml:"challenge_expiry"`
	ChallengeRateLimit float64       `koanf:"challenge_rate_limit" toml:"challenge_rate_limit"`
}

type AdminConfig struct {
	Enable        bool   `koanf:"enable" toml:"enable"`
	ListenAddress string `koanf:"listen" toml:"listen"`
	Pass          string `koanf:"pass" toml:"pass"`
	NoTLS         bool   `koanf:"notls" toml:"notls"`
	TLSCertFile   string `koanf:"cert" toml:"cert"`
	TLSKeyFile    string `koanf:"key" toml:"key"`
}

type SnapshotConfig struct {
	Enable          bool   `koanf:"enable" toml:"enable"`
	RecurringHeight uint64 `koanf:"recurring_height" toml:"recurring_height"`
	MaxSnapshots    uint64 `koanf:"max_snapshots" toml:"max_snapshots"`
}

type StateSyncConfig struct {
	Enable           bool     `koanf:"enable" toml:"enable"`
	TrustedProviders []string `koanf:"trusted_providers" toml:"trusted_providers"`

	DiscoveryTimeout time.Duration `koanf:"discovery_timeout" toml:"discovery_time"`
	MaxRetries       uint64        `koanf:"max_retries" toml:"max_retries"`
}

// ConfigToTOML marshals the config to TOML.
func (nc Config) ToTOML() ([]byte, error) {
	return toml.Marshal(nc)
}

func (nc *Config) FromTOML(b []byte) error {
	return toml.Unmarshal(b, &nc)
}

func (nc *Config) SaveAs(filename string) error {
	bts, err := nc.ToTOML()
	if err != nil {
		return err
	}

	// TODO: write a toml header/comment or perhaps use a text/template toml file

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
