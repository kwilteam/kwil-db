package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"

	"github.com/pelletier/go-toml/v2"
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
	// Leader is the leader's public key.
	Leader types.HexBytes `json:"leader"`
	// Validators is the list of genesis validators (including the leader).
	Validators []ktypes.Validator `json:"validators"`

	// TODO: more params like max block size, etc.
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
	PGConfig  PGConfig        `koanf:"pg" toml:"pg"`
	// RPC RPCConfig `koanf:"rpc" toml:"rpc"`
	// DB DBConfig `koanf:"db" toml:"db"`
}

// PeerConfig corresponds to the [peer] section of the config.
type PeerConfig struct {
	IP        string   `koanf:"ip" toml:"ip" comment:"ip to listen on for P2P connections"`
	Port      uint64   `koanf:"port" toml:"port" comment:"port to listen on for P2P connections"`
	Pex       bool     `koanf:"pex" toml:"pex" comment:"enable peer exchange"`
	BootNodes []string `koanf:"bootnodes" toml:"bootnodes" comment:"bootnodes to connect to on startup"`

	// ListenAddr string // "127.0.0.1:6600"
}

type PGConfig struct {
	// Host, Port, User, Pass, and DBName are used verbatim to create a
	// connection string in DSN format.
	// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
	Host           string `koanf:"host" toml:"host"`
	Port           string `koanf:"port" toml:"port"`
	User           string `koanf:"user" toml:"user"`
	Pass           string `koanf:"pass" toml:"pass"`
	DBName         string `koanf:"dbname" toml:"dbname"`
	MaxConnections uint32 `koanf:"max_connections" toml:"max_connections"`
}

type ConsensusConfig struct {
	ProposeTimeout time.Duration `koanf:"propose_timeout" toml:"propose_timeout" comment:"timeout for proposing a block"`
	MaxBlockSize   uint64        `koanf:"max_block_size" toml:"max_block_size" comment:"max size of a block in bytes"`
	MaxTxsPerBlock uint64        `koanf:"max_txs_per_block" toml:"max_txs_per_block" comment:"max number of transactions per block"`
	// ? reannounce intervals?
}

type RPCConfig struct {
	ListenAddress string        `koanf:"listen" toml:"listen"`
	Timeout       time.Duration `koanf:"timeout" toml:"timeout"`
	MaxReqSize    int           `koanf:"max_req_size" toml:"max_req_size"`
	// Private         bool          `koanf:"private" toml:"private"`
	// ChallengeExpiry    time.Duration `koanf:"challenge_expiry" toml:"challenge_expiry"`
	// ChallengeRateLimit float64       `koanf:"challenge_rate_limit" toml:"challenge_rate_limit"`
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
	Host          string        `koanf:"host" toml:"host"`
	Port          string        `koanf:"port" toml:"port"`
	User          string        `koanf:"user" toml:"user"`
	Pass          string        `koanf:"pass" toml:"pass"`
	DBName        string        `koanf:"dbname" toml:"dbname"`
	ReadTxTimeout time.Duration `koanf:"read_timeout" toml:"read_timeout"`
	MaxConns      uint32        `koanf:"max_connections" toml:"max_connections"`
}

func (nc *Config) SaveAs(filename string) error {
	bts, err := toml.Marshal(nc)
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
