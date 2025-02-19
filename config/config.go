package config

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"

	"github.com/pelletier/go-toml/v2"
)

var (
	ErrorExtraFields = errors.New("unrecognized fields")
)

const (
	ConfigFileName  = "config.toml"
	GenesisFileName = "genesis.json"

	DefaultAdminRPCAddr = "/tmp/kwild.socket"
	AdminServerKeyName  = "admin.key"
	AdminServerCertName = "admin.cert"

	MinProposeTimeout = types.Duration(500 * time.Millisecond)
)

type GenesisAlloc struct {
	ID      KeyHexBytes `json:"id"`
	KeyType string      `json:"key_type"`
	Amount  *big.Int    `json:"amount"`
}

// KeyHexBytes wraps hex bytes, and allows it to receive Ethereum 0x addresses
type KeyHexBytes struct{ types.HexBytes }

func (l *KeyHexBytes) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		if bytes.Equal(b, []byte("null")) {
			l.HexBytes = nil
			return nil
		}
		return fmt.Errorf("invalid hex string: %s", b)
	}
	sub := b[1 : len(b)-1] // strip the quotes

	// if it's an ethereum address, strip the 0x prefix
	if len(sub) == hex.EncodedLen(ethAddressLength)+2 && sub[0] == '0' && sub[1] == 'x' {
		sub = sub[2:]
	}

	dec := make([]byte, hex.DecodedLen(len(sub)))
	_, err := hex.Decode(dec, sub)
	if err != nil {
		return err
	}
	l.HexBytes = types.HexBytes(dec)
	return nil
}

const ethAddressLength = 20 // 20 bytes = 40 hex chars

func (l *KeyHexBytes) MarshalJSON() ([]byte, error) {
	if l.HexBytes == nil {
		return []byte("null"), nil
	}

	if len(l.HexBytes) == ethAddressLength {
		checksummed, err := auth.EthSecp256k1Authenticator{}.Identifier(l.HexBytes)
		if err != nil {
			return nil, err
		}

		return []byte(`"` + checksummed + `"`), nil
	}

	return []byte(`"` + hex.EncodeToString(l.HexBytes) + `"`), nil
}

type GenesisConfig struct {
	ChainID       string `json:"chain_id"`
	InitialHeight int64  `json:"initial_height"`

	// DBOwner is the owner of the database.
	// This should be either a public key or address.
	DBOwner string `json:"db_owner"`

	// Validators is the list of genesis validators (including the leader).
	Validators []*types.Validator `json:"validators"`

	// StateHash is the hash of the initial state of the chain, used when bootstrapping
	// the chain with a network snapshot during migration.
	StateHash types.HexBytes `json:"state_hash,omitempty"` // TODO: make it a *types.Hash
	// StateHash *types.Hash `json:"state_hash,omitempty"`

	// Alloc is the initial allocation of balances.
	Allocs []GenesisAlloc `json:"alloc,omitempty"`

	// Migration specifies the migration configuration required for zero downtime migration.
	Migration MigrationParams `json:"migration"`

	// NetworkParameters are network level configurations that can be
	// evolved over the lifetime of a network.
	types.NetworkParameters
}

func (gc *GenesisConfig) SanityChecks() error {
	switch len(gc.StateHash) {
	case 0, types.HashLen:
	default:
		return errors.New("invalid state hash, must be empty or 32 bytes")
	}

	if len(gc.Validators) == 0 {
		return errors.New("no validators provided")
	}

	// Validators power should not be zero
	for _, val := range gc.Validators {
		if val.Power <= 0 {
			return fmt.Errorf("Genesis validators should have non-zero power")
		}
	}

	if gc.InitialHeight < 0 {
		return errors.New("initial height must be greater than or equal to 0")
	}

	if err := gc.NetworkParameters.SanityChecks(); err != nil {
		return err
	}

	// Migration params should be both set or both unset
	if (gc.Migration.StartHeight == 0 && gc.Migration.EndHeight != 0) ||
		(gc.Migration.StartHeight != 0 && gc.Migration.EndHeight == 0) {
		return errors.New("both start and end height should be set or unset")
	}

	// ensure that the leader is part of the validator set
	isValidator := slices.ContainsFunc(gc.Validators, func(v *types.Validator) bool {
		if v.KeyType != gc.Leader.Type() {
			return false
		}
		return bytes.Equal(v.Identifier, gc.Leader.Bytes())
	})
	if !isValidator {
		return errors.New("leader is not part of the validator set")
	}

	return nil
}

func DecodePubKeyAndType(encodedPubKey string) ([]byte, crypto.KeyType, error) {
	parts := strings.Split(encodedPubKey, "#")
	if len(parts) != 2 {
		return nil, "", errors.New("invalid pubkey format, expected <pubkey#pubkeytype>")
	}

	pubKey, err := hex.DecodeString(parts[0])
	if err != nil {
		return nil, "", fmt.Errorf("error decoding public key: %s error: %s", parts[0], err)
	}

	pubKeyType := parts[1]
	if pubKeyType == "" || strings.ContainsRune(pubKeyType, ' ') {
		return nil, "", errors.New("invalid pubkey type, expected <pubkey#pubkeytype>")
	}

	return pubKey, crypto.KeyType(pubKeyType), nil
}

func EncodePubKeyAndType(pubKey []byte, pubKeyType crypto.KeyType) string {
	return fmt.Sprintf("%s#%s", hex.EncodeToString(pubKey), pubKeyType)
}

func FormatAccountID(acctID *types.AccountID) string {
	if acctID == nil {
		return ""
	}
	return fmt.Sprintf("%s#%s", acctID.Identifier.String(), acctID.KeyType)
}

// MigrationParams is the migration configuration required for zero downtime
// migration. The height values refer to the height of the old/from chain.
type MigrationParams struct {
	// StartHeight is the height from which the state from the old chain is to be migrated.
	StartHeight int64 `json:"start_height"`
	// EndHeight is the height till which the state from the old chain is to be migrated.
	EndHeight int64 `json:"end_height"`
}

func (m *MigrationParams) IsMigration() bool {
	return m.StartHeight != 0 || m.EndHeight != 0
}

func (nc *GenesisConfig) SaveAs(filename string) error {
	bts, err := json.MarshalIndent(nc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, bts, 0644)
}

func LoadGenesisConfig(filename string) (*GenesisConfig, error) {
	fid, err := os.Open(filename)
	if err != nil {
		return nil, err // can be os.ErrNotExist
	}
	defer fid.Close()

	var nc GenesisConfig
	dec := json.NewDecoder(fid)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&nc); err != nil {
		return nil, err
	}

	return &nc, nil
}

func DefaultGenesisConfig() *GenesisConfig {
	return &GenesisConfig{
		ChainID:       "kwil-test-chain",
		InitialHeight: 0,
		DBOwner:       "",
		Validators:    nil,
		StateHash:     nil,
		Migration:     MigrationParams{},
		NetworkParameters: types.NetworkParameters{
			Leader:           types.PublicKey{ /* nil crypto.PublicKey */ },
			MaxBlockSize:     6 * 1024 * 1024,
			JoinExpiry:       types.Duration(7 * 24 * time.Hour), // 1 week
			DisabledGasCosts: true,
			MaxVotesPerTx:    200,
			MigrationStatus:  types.NoActiveMigration,
		},
	}
}

// DefaultConfig generates an instance of the default config.
func DefaultConfig() *Config {
	return &Config{
		LogLevel:  log.LevelInfo,
		LogFormat: log.FormatUnstructured,
		LogOutput: []string{"stdout", "kwild.log"},
		P2P: PeerConfig{
			ListenAddress:     "0.0.0.0:6600",
			Pex:               true,
			BootNodes:         []string{},
			TargetConnections: 20,
		},
		Consensus: ConsensusConfig{
			ProposeTimeout:        types.Duration(1000 * time.Millisecond),
			EmptyBlockTimeout:     types.Duration(1 * time.Minute),
			BlockProposalInterval: types.Duration(1 * time.Second),
			BlockAnnInterval:      types.Duration(3 * time.Second),
		},
		Store: StoreConfig{
			Compression: true,
		},
		DB: DBConfig{
			Host:          "127.0.0.1",
			Port:          "5432",
			User:          "kwild",
			Pass:          "",
			DBName:        "kwild",
			ReadTxTimeout: types.Duration(45 * time.Second),
			MaxConns:      60,
		},
		RPC: RPCConfig{
			ListenAddress:      "0.0.0.0:8484",
			BroadcastTxTimeout: types.Duration(15 * time.Second),
			Timeout:            types.Duration(20 * time.Second),
			MaxReqSize:         6_000_000,
			Private:            false,
			ChallengeExpiry:    types.Duration(30 * time.Second),
			ChallengeRateLimit: 10,
		},
		Admin: AdminConfig{
			Enable:        true,
			ListenAddress: DefaultAdminRPCAddr,
			Pass:          "",
			NoTLS:         false,
			// TLSCertFile:   AdminCertName,
			// TLSKeyFile:    "admin.key",
		},
		Snapshots: SnapshotConfig{
			Enable:          false,
			RecurringHeight: 14400,
			MaxSnapshots:    3,
		},
		StateSync: StateSyncConfig{
			Enable:           false,
			DiscoveryTimeout: types.Duration(15 * time.Second),
			MaxRetries:       3,
		},
		Extensions: make(map[string]map[string]string),
		Checkpoint: Checkpoint{
			Height: 0,
			Hash:   types.Hash{},
		},
		Erc20BridgeSigner: ERC20BridgeSignerConfig{
			Enable:      false,
			PrivateKeys: nil,
			Targets:     nil,
			// the reasonable value is the block time
			SyncEvery: types.Duration(1 * time.Minute),
		},
	}
}

// Config is the node's config.
type Config struct {
	LogLevel  log.Level  `toml:"log_level" comment:"log level\npossible values: 'debug', 'info', 'warn', and 'error'"`
	LogFormat log.Format `toml:"log_format" comment:"log format\npossible values: 'json', 'text' (kv), and 'plain' (fmt-style)"`
	LogOutput []string   `toml:"log_output" comment:"output paths for the log"`

	ProfileMode string `toml:"profile_mode,commented" comment:"profile mode (http, cpu, mem, mutex, or block)"`
	ProfileFile string `toml:"profile_file,commented" comment:"profile output file path (e.g. cpu.pprof)"`

	P2P               PeerConfig                   `toml:"p2p" comment:"P2P related configuration"`
	Consensus         ConsensusConfig              `toml:"consensus" comment:"Consensus related configuration"`
	DB                DBConfig                     `toml:"db" comment:"DB (PostgreSQL) related configuration"`
	Store             StoreConfig                  `toml:"store" comment:"Block store configuration"`
	RPC               RPCConfig                    `toml:"rpc" comment:"User RPC service configuration"`
	Admin             AdminConfig                  `toml:"admin" comment:"Admin RPC service configuration"`
	Snapshots         SnapshotConfig               `toml:"snapshots" comment:"Snapshot creation and provider configuration"`
	StateSync         StateSyncConfig              `toml:"state_sync" comment:"Statesync configuration (vs block sync)"`
	Extensions        map[string]map[string]string `toml:"extensions" comment:"extension configuration"`
	GenesisState      string                       `toml:"genesis_state" comment:"path to the genesis state file, relative to the root directory"`
	Migrations        MigrationConfig              `toml:"migrations" comment:"zero downtime migration configuration"`
	Checkpoint        Checkpoint                   `toml:"checkpoint" comment:"checkpoint info for the leader to sync to before proposing a new block"`
	Erc20BridgeSigner ERC20BridgeSignerConfig      `toml:"erc20_bridge_signer" comment:"ERC20 bridge signer service configuration"`
}

// PeerConfig corresponds to the [p2p] section of the config.
type PeerConfig struct {
	ListenAddress     string   `toml:"listen" comment:"address in host:port format to listen on for P2P connections"`
	Pex               bool     `toml:"pex" comment:"enable peer exchange"`
	BootNodes         []string `toml:"bootnodes" comment:"bootnodes to connect to on startup"`
	PrivateMode       bool     `toml:"private" comment:"operate in private mode using a node ID whitelist"`
	Whitelist         []string `toml:"whitelist" comment:"allowed node IDs when in private mode"`
	TargetConnections int      `toml:"target_connections" comment:"target number of connections to maintain"`
	ExternalAddress   string   `toml:"external_address" comment:"external address in host:port format to advertise to the network"`
}

// StoreConfig contains options related to the block store. This is the embedded
// database used to store the raw block data, unlike the DBConfig which is
// effectively the state store.
type StoreConfig struct {
	Compression bool `toml:"compression" comment:"compress data when writing new data"`

	// Internal block size and block cache size may be of use soon.
	//   https://github.com/kwilteam/kwil-db/issues/1347
	// CacheSize int `toml:"cache_size" comment:"size of the block store cache in bytes"`
	// ChunkSize int `toml:"chunk_size" comment:"size of the block store's internal blocks"`
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
	Host          string         `toml:"host" comment:"postgres host name (IP or UNIX socket path)"`
	Port          string         `toml:"port" comment:"postgres TCP port (leave empty for UNIX socket)"`
	User          string         `toml:"user" comment:"postgres role/user name"`
	Pass          string         `toml:"pass" comment:"postgres password if required for the user and host"`
	DBName        string         `toml:"dbname" comment:"postgres database name"`
	ReadTxTimeout types.Duration `toml:"read_timeout" comment:"timeout on read transactions from user RPC calls and queries"`
	MaxConns      uint32         `toml:"max_connections" comment:"maximum number of DB connections to permit"`
}

type ConsensusConfig struct {
	ProposeTimeout types.Duration `toml:"propose_timeout" comment:"minimum duration to wait before proposing a block with transactions (applies to leader). This value should be greater than 500ms."`

	EmptyBlockTimeout types.Duration `toml:"empty_block_timeout" comment:"timeout for proposing an empty block. If set to 0, disables empty blocks, leader will wait indefinitely until transactions are available to produce a block."`

	// BlockProposalInterval is the interval between block proposal reannouncements by the leader.
	// This affects the time it takes for an out-of-sync validator to receive the current block proposal,
	// thereby impacting the block times. Default is 1 second.
	BlockProposalInterval types.Duration `toml:"block_proposal_interval" comment:"interval between block proposal reannouncements by the leader"`

	// BlockAnnInterval is the frequency with which the block commit messages are reannounced by the leader,
	// and votes reannounced by validators. Default is 3 seconds. This affects the time it takes for
	// out-of-sync nodes to catch up with the latest block.
	BlockAnnInterval types.Duration `toml:"block_ann_interval" comment:"interval between block commit reannouncements by the leader, and votes reannouncements by validators"`
}

type RPCConfig struct {
	ListenAddress      string         `toml:"listen" comment:"address in host:port format on which the RPC server will listen"`
	BroadcastTxTimeout types.Duration `toml:"broadcast_tx_timeout" comment:"duration to wait for a tx to be committed when transactions are authored with --sync flag"`
	Timeout            types.Duration `toml:"timeout" comment:"user request duration limit after which it is cancelled"`
	MaxReqSize         int            `toml:"max_req_size" comment:"largest permissible user request size"`
	Private            bool           `toml:"private" comment:"enable private mode that requires challenge authentication for each call"`
	Compression        bool           `toml:"compression" comment:"use compression in RPC responses"`
	ChallengeExpiry    types.Duration `toml:"challenge_expiry" comment:"lifetime of a server-generated challenge"`
	ChallengeRateLimit float64        `toml:"challenge_rate_limit" comment:"maximum number of challenges per second that a user can request"`
}

type AdminConfig struct {
	Enable        bool   `toml:"enable" comment:"enable the admin RPC service"`
	ListenAddress string `toml:"listen" comment:"address in host:port format or UNIX socket path on which the admin RPC server will listen"`
	Pass          string `toml:"pass" comment:"optional password for the admin service"`
	NoTLS         bool   `toml:"notls" comment:"disable TLS when the listen address is not a loopback IP or UNIX socket"`
	// TLSCertFile   string `toml:"cert" comment:"TLS certificate for use with a non-loopback listen address when notls is not true"`
	// TLSKeyFile    string `toml:"key" comment:"TLS key for use with a non-loopback listen address when notls is not true"`
}

type SnapshotConfig struct {
	Enable          bool   `toml:"enable" comment:"enable creating and providing snapshots for peers using statesync"`
	RecurringHeight uint64 `toml:"recurring_height" comment:"snapshot creation period in blocks"`
	MaxSnapshots    uint64 `toml:"max_snapshots" comment:"number of snapshots to keep, after the oldest is removed when creating a new one"`
}

type StateSyncConfig struct {
	Enable           bool     `toml:"enable" comment:"enable using statesync rather than blocksync"`
	TrustedProviders []string `toml:"trusted_providers" comment:"trusted snapshot providers in node ID format (see bootnodes)"`

	DiscoveryTimeout types.Duration `toml:"discovery_time" comment:"how long to discover snapshots before selecting one to use"`
	MaxRetries       uint64         `toml:"max_retries" comment:"how many times to try after failing to apply a snapshot before switching to blocksync"`
}

type MigrationConfig struct {
	Enable      bool   `toml:"enable" comment:"enable zero downtime migrations"`
	MigrateFrom string `toml:"migrate_from" comment:"JSON-RPC listening address of the node to replicate the state from"`
}

type Checkpoint struct {
	// Height 0 indicates no checkpoint is set. The leader will attempt regular block sync.
	Height int64      `toml:"height" comment:"checkpoint height for the leader. If the leader is behind this height, it will sync to this height before attempting to propose a new block."`
	Hash   types.Hash `toml:"hash" comment:"checkpoint block hash."`
}

type ERC20BridgeSignerConfig struct {
	Enable      bool           `toml:"enable" comment:"enable the ERC20 bridge signer service"`
	Targets     []string       `toml:"targets" comment:"target reward ext alias for the ERC20 reward"`
	PrivateKeys []string       `toml:"private_keys" comment:"private key for the ERC20 reward target"`
	EthRpcs     []string       `toml:"eth_rpcs" comment:"eth rpc address for the ERC20 reward target"`
	SyncEvery   types.Duration `toml:"sync_every" comment:"sync interval; a recommend value is same as the block time"`
}

func (cfg ERC20BridgeSignerConfig) Validate() error {
	if (len(cfg.PrivateKeys) != len(cfg.Targets)) && (len(cfg.EthRpcs) != len(cfg.Targets)) {
		return fmt.Errorf("private keys and targets and eth_rpcs must be configured in triples")
	}

	if len(cfg.Targets) == 0 {
		return fmt.Errorf("no target configured")
	}

	for i, target := range cfg.Targets {
		if target == "" {
			return fmt.Errorf("target %dth is empty", i)
		}

		if cfg.PrivateKeys[i] == "" {
			return fmt.Errorf("private key %dth is empty", i)
		}

		if cfg.EthRpcs[i] == "" {
			return fmt.Errorf("eth rpc %dth is empty", i)
		}
	}

	return nil
}

// ToTOML marshals the config to TOML. The `toml` struct field tag
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
	return nc.fromTOML(bytes.NewReader(b))
}

func (nc *Config) fromTOML(rd io.Reader) error {
	dec := toml.NewDecoder(rd)
	dec.DisallowUnknownFields()
	err := dec.Decode(&nc)
	var tomlErr *toml.StrictMissingError
	if errors.As(err, &tomlErr) {
		err = fmt.Errorf("%w:\n%s", ErrorExtraFields, tomlErr.String())
	}
	return err
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
	fid, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fid.Close()

	var nc Config
	if err := nc.fromTOML(fid); err != nil {
		return nil, err
	}

	return &nc, nil
}
