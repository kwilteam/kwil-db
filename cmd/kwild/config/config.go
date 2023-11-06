// Package config provides types and functions for node configuration loading
// and generation.
package config

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"

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
	AutoGen bool

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
	GrpcListenAddress  string   `mapstructure:"grpc_listen_addr"`
	HTTPListenAddress  string   `mapstructure:"http_listen_addr"`
	AdminListenAddress string   `mapstructure:"admin_listen_addr"`
	PrivateKeyPath     string   `mapstructure:"private_key_path"`
	SqliteFilePath     string   `mapstructure:"sqlite_file_path"`
	ExtensionEndpoints []string `mapstructure:"extension_endpoints"`
	//SnapshotConfig     SnapshotConfig `mapstructure:"snapshots"`
	TLSCertFile  string `mapstructure:"tls_cert_file"`
	TLSKeyFile   string `mapstructure:"tls_key_file"`
	EnableRPCTLS bool   `mapstructure:"rpctls"`
	Hostname     string `mapstructure:"hostname"`
	ProfileMode  string `mapstructure:"profile_mode"`
	ProfileFile  string `mapstructure:"profile_file"`
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
	HandshakeTimeout time.Duration `mapstructure:"handshake_timeout"`
	// DialTimeout is the peer connection establishment timeout.
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
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
	TimeoutPropose time.Duration `mapstructure:"timeout_propose"`
	// TimeoutPrevote is how long to wait after receiving +2/3 prevotes for
	// “anything” (i.e. not a single block or nil).
	TimeoutPrevote time.Duration `mapstructure:"timeout_prevote"`
	// TimeoutPrecommit is how long we wait after receiving +2/3 precommits for
	// “anything” (i.e. not a single block or nil).
	TimeoutPrecommit time.Duration `mapstructure:"timeout_precommit"`
	// TimeoutCommit is how long to wait after committing a block, before
	// starting on the new height (this gives us a chance to receive some more
	// precommits, even though we already have +2/3).
	TimeoutCommit time.Duration `mapstructure:"timeout_commit"`
}

type StateSyncConfig struct {
	Enable              bool          `mapstructure:"enable"`
	TempDir             string        `mapstructure:"temp_dir"`
	RPCServers          []string      `mapstructure:"rpc_servers"`
	DiscoveryTime       time.Duration `mapstructure:"discovery_time"`
	ChunkRequestTimeout time.Duration `mapstructure:"chunk_request_timeout"`
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

func defaultMoniker() string {
	moniker, err := os.Hostname()
	if err != nil {
		moniker = "amnesiac"
	}
	return moniker
}

func (cfg *KwildConfig) LoadKwildConfig() error {
	var err error
	cfg.RootDir, err = ExpandPath(cfg.RootDir)
	if err != nil {
		return fmt.Errorf("failed to expand root directory \"%v\": %v", cfg.RootDir, err)
	}

	fmt.Printf("kwild starting with root directory \"%v\"\n", cfg.RootDir)

	cfgFile := filepath.Join(cfg.RootDir, ConfigFileName)
	err = cfg.ParseConfig(cfgFile) // viper magic here
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	cfg.sanitizeCfgPaths()
	cfg.configureCerts()

	if cfg.ChainCfg.Moniker == "" {
		cfg.ChainCfg.Moniker = defaultMoniker()
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (cfg *KwildConfig) ParseConfig(cfgFile string) error {
	/*
		Lots of Viper magic here, but the gist is:
		We want to be able to set config values via
			-  flags
			-  environment variables
			-  config file
			-  default values

		for env variables support:
		Requirement is, we need to be able to config from env variables with a prefix "KWILD_"

		It can be done 2 ways:
		1. AutomaticEnv: off mode
			- This will not bind env variables to config values automatically
			- We need to manually bind env variables to config values (this is what we are doing currently)
			- As we bound flags to viper, viper is already aware of the config structure mapping,
				so we can explicitly call viper.BindEnv() on all the keys in viper.AllKeys()
			- else we would have to reflect on the config structure and bind env variables to config values

		2. AutomaticEnv: on mode
			- This is supposed to automatically bind env variables to config values
				(but it doesn't work without doing a bit more work from our side)
			- One way to make this work is add default values using either viper.SetDefault() for all the config values
			  or can do viper.MergeConfig(<serialized config>)
			- Serializing is really painful as cometbft has a field which is using map<interface><interface> though its deprecated.
				which prevents us from doing the AutomaticEnv binding
		Issues referencing the issues (or) correct usage of AutomaticEnv: https://github.com/spf13/viper/issues/188
		For now, we are going with the first approach

		Note:
		The order of preference of various modes of config supported by viper is:
		explicit call to Set > flags > env variables > config file > default values
	*/
	for _, key := range viper.AllKeys() {
		envKey := "KWILD_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		viper.BindEnv(key, envKey)
	}

	if fileExists(cfgFile) {
		fmt.Println("Loading config from: ", cfgFile)
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("reading config: %v", err)
		}
	} else {
		fmt.Printf("Config file %s not found. Using default settings.\n", cfgFile)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("decoding config: %v", err)
	}
	return nil
}

func DefaultConfig() *KwildConfig {
	return &KwildConfig{
		AppCfg: &AppConfig{
			GrpcListenAddress:  "localhost:50051",
			HTTPListenAddress:  "localhost:8080",
			AdminListenAddress: "localhost:50151",
			SqliteFilePath:     DefaultSQLitePath,
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
				AllowDuplicateIP:    true,  // override comet
				PexReactor:          false, // override comet - not recommended for validators
				HandshakeTimeout:    20 * time.Second,
				DialTimeout:         3 * time.Second,
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
				DiscoveryTime:       15 * time.Second,
				ChunkRequestTimeout: 10 * time.Second,
			},
			Consensus: &ConsensusConfig{
				TimeoutPropose:   3 * time.Second,
				TimeoutPrevote:   2 * time.Second,
				TimeoutPrecommit: 2 * time.Second,
				TimeoutCommit:    6 * time.Second,
			},
		},
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
	cfg.AppCfg.SqliteFilePath = rootify(cfg.AppCfg.SqliteFilePath, rootDir)
	//cfg.AppCfg.SnapshotConfig.SnapshotDir = rootify(cfg.AppCfg.SnapshotConfig.SnapshotDir, rootDir)

	if cfg.AppCfg.PrivateKeyPath != "" {
		cfg.AppCfg.PrivateKeyPath = rootify(cfg.AppCfg.PrivateKeyPath, rootDir)
	} else {
		cfg.AppCfg.PrivateKeyPath = filepath.Join(rootDir, PrivateKeyFileName)
	}

	fmt.Println("Private key path:", cfg.AppCfg.PrivateKeyPath)
}

func (cfg *KwildConfig) InitPrivateKeyAndGenesis() (privateKey *crypto.Ed25519PrivateKey, genConfig *GenesisConfig, err error) {
	return loadGenesisAndPrivateKey(cfg.AutoGen, cfg.AppCfg.PrivateKeyPath, cfg.RootDir)
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
