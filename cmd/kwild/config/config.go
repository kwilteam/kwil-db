// Package config provides types and functions for node configuration loading
// and generation.
package config

import (
	"bytes"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"

	merge "dario.cat/mergo"
	"github.com/mitchellh/mapstructure"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	DefaultSnapshotsDir = "snapshots"

	DefaultTLSCertFile  = "rpc.cert"
	defaultTLSKeyFile   = "rpc.key"
	defaultAdminClients = "clients.pem"
)

var DefaultSQLitePath = filepath.Join("data", "kwild.db") // a folder, not a file

type KwildConfig struct {
	RootDir string

	AppCfg   *AppConfig   `mapstructure:"app"`
	ChainCfg *ChainConfig `mapstructure:"chain"`
	Logging  *Logging     `mapstructure:"log"`
}

type Logging struct {
	Level        string   `mapstructure:"level"`
	Format       string   `mapstructure:"format"`
	TimeEncoding string   `mapstructure:"time_format"`
	OutputPaths  []string `mapstructure:"output_paths"`
}

type AppConfig struct {
	GrpcListenAddress  string `mapstructure:"grpc_listen_addr"`
	HTTPListenAddress  string `mapstructure:"http_listen_addr"`
	AdminListenAddress string `mapstructure:"admin_listen_addr"`
	PrivateKeyPath     string `mapstructure:"private_key_path"`

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

	ExtensionEndpoints []string                     `mapstructure:"extension_endpoints"`
	TLSCertFile        string                       `mapstructure:"tls_cert_file"`
	TLSKeyFile         string                       `mapstructure:"tls_key_file"`
	EnableRPCTLS       bool                         `mapstructure:"rpctls"`
	Hostname           string                       `mapstructure:"hostname"`
	ProfileMode        string                       `mapstructure:"profile_mode"`
	ProfileFile        string                       `mapstructure:"profile_file"`
	Oracles            map[string]map[string]string `mapstructure:"oracles"`

	// SnapshotConfig     SnapshotConfig `mapstructure:"snapshots"`
}

type SnapshotConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	RecurringHeight uint64 `mapstructure:"snapshot_heights"`
	MaxSnapshots    uint64 `mapstructure:"max_snapshots"`
	SnapshotDir     string `mapstructure:"snapshot_dir"`
}

type ChainRPCConfig struct {
	// TCP or UNIX socket address for the RPC server to listen on
	ListenAddress string `mapstructure:"listen_addr"`
}

type P2PConfig struct {
	// ListenAddress is the address on which to listen for incoming connections.
	ListenAddress string `mapstructure:"listen_addr"`
	// ExternalAddress is the address to advertise to peers to dial us.
	ExternalAddress string `mapstructure:"external_address"`
	// PersistentPeers is a comma separated list of nodes to keep persistent
	// connections to.
	PersistentPeers string `mapstructure:"persistent_peers"`
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
	Enable              bool     `mapstructure:"enable"`
	TempDir             string   `mapstructure:"temp_dir"`
	RPCServers          []string `mapstructure:"rpc_servers"`
	DiscoveryTime       Duration `mapstructure:"discovery_time"`
	ChunkRequestTimeout Duration `mapstructure:"chunk_request_timeout"`
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

var _ encoding.TextUnmarshaler = (*Duration)(nil)
var _ pflag.Value = (*Duration)(nil)

func (d *Duration) UnmarshalText(b []byte) error {
	x, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

func (d *Duration) String() string {
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

func defaultMoniker() string {
	moniker, err := os.Hostname()
	if err != nil {
		moniker = "amnesiac"
	}
	return moniker
}

// GetCfg gets the kwild config
// It has the following precedence (low to high):
// 1. Default
// 2. Config file
// 3. Env vars
// 4. Command line flags
// It takes one argument, which is the config generated from the command line flags.
func GetCfg(flagCfg *KwildConfig) (*KwildConfig, error) {
	/*
		the process here is:
		1. identify the root dir.  This requires reading in the env and command line flags
		to see if they specify a root dir (since they take precedence over the config file).
		If no root dir is specified from these, then use the default root dir.
		2. Read in the config file, if it exists, and merge it into the default config.
		3. Merge in the env config.
		4. Merge in the flag config.
	*/

	// 1. identify the root dir
	cfg := DefaultConfig()
	rootDir := cfg.RootDir

	// read in env config
	envCfg, err := LoadEnvConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load env config: %w", err)
	}
	if envCfg.RootDir != "" {
		rootDir = envCfg.RootDir
	}

	if flagCfg.RootDir != "" {
		rootDir = flagCfg.RootDir
	}

	// expand the root dir
	rootDir, err = ExpandPath(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to expand root directory \"%v\": %v", rootDir, err)
	}

	fmt.Printf("Root directory \"%v\"\n", rootDir)

	// make sure the root dir exists
	err = os.MkdirAll(rootDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create root directory \"%v\": %v", rootDir, err)
	}

	// 2. Read in the config file
	// read in config file and merge into default config
	fileCfg, err := LoadConfigFile(filepath.Join(rootDir, ConfigFileName))
	if err == nil {
		// merge in config file
		err2 := cfg.Merge(fileCfg)
		if err2 != nil {
			return nil, fmt.Errorf("failed to merge config file: %w", err2)
		}
	} else if err != ErrConfigFileNotFound {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// 3. Merge in the env config
	// merge in env config
	err = cfg.Merge(envCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to merge env config: %w", err)
	}

	// 4. Merge in the flag config
	// merge in flag config
	err = cfg.Merge(flagCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to merge flag config: %w", err)
	}

	cfg.RootDir = rootDir

	cfg.sanitizeCfgPaths()
	cfg.configureCerts()
	if cfg.ChainCfg.Moniker == "" {
		cfg.ChainCfg.Moniker = defaultMoniker()
	}

	return cfg, nil
}

// LoadConfig reads a config.toml at the given path and returns a KwilConfig.
// If the file does not exist, it will return an ErrConfigFileNotFound error.
func LoadConfigFile(configPath string) (*KwildConfig, error) {
	cfgFilePath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of config file: %v due to error: %v", configPath, err)
	}

	if !fileExists(cfgFilePath) {
		return nil, ErrConfigFileNotFound
	}

	bts, err := os.ReadFile(cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// unmarshal toml to maps
	var cfg map[string]interface{}
	err = toml.Unmarshal(bts, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// convert mapstructure toml to KwilConfig
	var kwilCfg KwildConfig

	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			// func to decode string to Duration
			func(
				f reflect.Type,
				t reflect.Type,
				data interface{}) (interface{}, error) {
				if f.Kind() != reflect.String {
					return data, nil
				}
				if t != reflect.TypeOf(Duration(time.Duration(5))) {
					return data, nil
				}

				// Convert it by parsing
				dur, err := time.ParseDuration(data.(string))
				if err != nil {
					return nil, err
				}

				return Duration(dur), nil
			},
			// func to decode string to []string{} if the field is of type []string
			// AFAICT this is only used for statesync rpc servers, which while not released,
			// we do have some tooling for it
			func(
				f reflect.Type,
				t reflect.Type,
				data interface{}) (interface{}, error) {
				if f.Kind() != reflect.String {
					return data, nil
				}

				if t != reflect.TypeOf([]string{}) {
					return data, nil
				}

				// parse comma separated string to []string
				return strings.Split(data.(string), ","), nil
			},
		),
		Result: &kwilCfg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create mapstructure decoder: %v", err)
	}

	err = mapDecoder.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config file: %v", err)
	}

	return &kwilCfg, nil
}

// LoadEnvConfig loads a config from environment variables.
func LoadEnvConfig() (*KwildConfig, error) {
	viper.SetEnvPrefix("KWILD")
	viper.AutomaticEnv()

	var cfg KwildConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decoding config: %v", err)
	}

	return &cfg, nil
}

var ErrConfigFileNotFound = fmt.Errorf("config file not found")

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DefaultConfig() *KwildConfig {
	return &KwildConfig{
		AppCfg: &AppConfig{
			GrpcListenAddress:  "localhost:50051",
			HTTPListenAddress:  "localhost:8080",
			AdminListenAddress: "unix:///tmp/kwil_admin.sock",
			// SqliteFilePath:     DefaultSQLitePath,
			DBHost: "/var/run/postgresql",
			DBPort: "5432", // ignored with unix socket, but applies if IP used for DBHost
			DBUser: "kwild",
			DBName: "kwild",
			// SnapshotConfig: SnapshotConfig{
			// 	Enabled:         false,
			// 	RecurringHeight: uint64(10000),
			// 	MaxSnapshots:    3,
			// 	SnapshotDir:     DefaultSnapshotsDir,
			// },
		},
		Logging: &Logging{
			Level:        "info",
			Format:       log.FormatJSON,
			TimeEncoding: log.TimeEncodingEpochFloat,
			OutputPaths:  []string{"stdout"},
		},
		ChainCfg: &ChainConfig{
			P2P: &P2PConfig{
				ListenAddress:       "tcp://0.0.0.0:26656",
				ExternalAddress:     "",
				AddrBookStrict:      false, // override comet
				MaxNumInboundPeers:  40,
				MaxNumOutboundPeers: 10,
				AllowDuplicateIP:    true, // override comet
				PexReactor:          true,
				HandshakeTimeout:    Duration(20 * time.Second),
				DialTimeout:         Duration(3 * time.Second),
			},
			RPC: &ChainRPCConfig{
				ListenAddress: "tcp://127.0.0.1:26657",
			},
			Mempool: &MempoolConfig{
				Size:        5000,
				CacheSize:   10000,
				MaxTxBytes:  1024 * 1024 * 4,   // 4 MiB
				MaxTxsBytes: 1024 * 1024 * 512, // 512 MiB
			},
			StateSync: &StateSyncConfig{
				Enable:              false,
				DiscoveryTime:       Duration(15 * time.Second),
				ChunkRequestTimeout: Duration(10 * time.Second),
			},
			Consensus: &ConsensusConfig{
				TimeoutPropose:   Duration(3 * time.Second),
				TimeoutPrevote:   Duration(2 * time.Second),
				TimeoutPrecommit: Duration(2 * time.Second),
				TimeoutCommit:    Duration(6 * time.Second),
			},
		},
	}
}

// EmptyConfig returns a config with all fields set to their zero values.
// This is useful for guaranteeing that all fields are set when merging
func EmptyConfig() *KwildConfig {
	return &KwildConfig{
		AppCfg: &AppConfig{
			ExtensionEndpoints: []string{},
		},
		ChainCfg: &ChainConfig{
			P2P:     &P2PConfig{},
			RPC:     &ChainRPCConfig{},
			Mempool: &MempoolConfig{},
			StateSync: &StateSyncConfig{
				RPCServers: []string{},
			},
			Consensus: &ConsensusConfig{},
		},
		Logging: &Logging{},
	}
}

func (cfg *KwildConfig) LogConfig() *log.Config {
	// Rootify any relative paths.
	outputPaths := make([]string, 0, len(cfg.Logging.OutputPaths))
	for _, path := range cfg.Logging.OutputPaths {
		switch path {
		case "stdout", "stderr":
			outputPaths = append(outputPaths, path)
		default:
			outputPaths = append(outputPaths, rootify(path, cfg.RootDir))
		}
	}
	// log.Config <== config.Logging
	return &log.Config{
		Level:       cfg.Logging.Level,
		OutputPaths: outputPaths,
		Format:      cfg.Logging.Format,
		EncodeTime:  cfg.Logging.TimeEncoding,
	}
}

func (cfg *KwildConfig) configureCerts() {
	if cfg.AppCfg.TLSCertFile == "" {
		cfg.AppCfg.TLSCertFile = DefaultTLSCertFile
	}
	cfg.AppCfg.TLSCertFile = rootify(cfg.AppCfg.TLSCertFile, cfg.RootDir)

	if cfg.AppCfg.TLSKeyFile == "" {
		cfg.AppCfg.TLSKeyFile = defaultTLSKeyFile
	}
	cfg.AppCfg.TLSKeyFile = rootify(cfg.AppCfg.TLSKeyFile, cfg.RootDir)
}

func (cfg *KwildConfig) sanitizeCfgPaths() {
	rootDir := cfg.RootDir
	//cfg.AppCfg.SnapshotConfig.SnapshotDir = rootify(cfg.AppCfg.SnapshotConfig.SnapshotDir, rootDir)

	if cfg.AppCfg.PrivateKeyPath != "" {
		cfg.AppCfg.PrivateKeyPath = rootify(cfg.AppCfg.PrivateKeyPath, rootDir)
	} else {
		cfg.AppCfg.PrivateKeyPath = filepath.Join(rootDir, PrivateKeyFileName)
	}

	fmt.Println("Private key path:", cfg.AppCfg.PrivateKeyPath)
}

func (cfg *KwildConfig) InitPrivateKeyAndGenesis(autogen bool) (privateKey *crypto.Ed25519PrivateKey, genConfig *GenesisConfig, err error) {
	return loadGenesisAndPrivateKey(autogen, cfg.AppCfg.PrivateKeyPath, cfg.RootDir)
}

func ExpandPath(path string) (string, error) {
	var expandedPath string
	if tail, cut := strings.CutPrefix(path, "~/"); cut {
		// Expands ~/ in the path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		expandedPath = filepath.Join(homeDir, tail)
	} else {
		// Expands relative paths
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path of file: %v due to error: %v", path, err)
		}
		expandedPath = absPath
	}
	return expandedPath, nil
}

// saveNodeKey writes the private key hexadecimal encoded to a file.
func saveNodeKey(priv []byte, keyPath string) error {
	keyHex := hex.EncodeToString(priv[:])
	return os.WriteFile(keyPath, []byte(keyHex), 0600)
}

// loadNodeKey loads a Kwil node private key file.
func loadNodeKey(keyFile string) (priv, pub []byte, err error) {
	privKeyHexB, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading private key file: %w", err)
	}
	privKeyHex := string(bytes.TrimSpace(privKeyHexB))
	privB, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding private key: %w", err)
	}
	privKey, err := crypto.Ed25519PrivateKeyFromBytes(privB)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private key: %w", err)
	}
	pubKey := privKey.PubKey()
	return privKey.Bytes(), pubKey.Bytes(), nil
}

// newNodeKey generates a node key pair, returning both as bytes.
func newNodeKey() (priv, pub []byte, err error) {
	privKey, err := crypto.GenerateEd25519Key()
	if err != nil {
		return nil, nil, err
	}
	return privKey.Bytes(), privKey.PubKey().Bytes(), nil
}

// ReadOrCreatePrivateKeyFile will read the node key pair from the given file,
// or generate it if it does not exist and requested.
func ReadOrCreatePrivateKeyFile(keyPath string, autogen bool) (priv, pub []byte, generated bool, err error) {
	priv, pub, err = loadNodeKey(keyPath)
	if err == nil {
		return priv, pub, false, nil
	}

	if !autogen {
		return nil, nil, false, fmt.Errorf("private key not found")
	}

	priv, pub, err = newNodeKey()
	if err != nil {
		return nil, nil, false, err
	}

	return priv, pub, true, saveNodeKey(priv, keyPath)
}
