// Package config provides types and functions for node configuration loading
// and generation.
package config

import (
	"bytes"
	"encoding"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/mitchellh/mapstructure"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
)

const (
	DefaultTLSCertFile  = "rpc.cert"
	defaultTLSKeyFile   = "rpc.key"
	defaultAdminClients = "clients.pem"
)

var _ encoding.TextUnmarshaler = (*config.Duration)(nil)

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
// It takes the config generated from the command line flags to override default.
// It also takes a flag to indicate if the caller wants to modify the defaults
// for "quickstart" mode. Presently this just makes the HTTP RPC service listen
// on all interfaces instead of the default of localhost.
func GetCfg(flagCfg *config.KwildConfig) (*config.KwildConfig, bool, error) {
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
	cfg := cmd.DefaultConfig()
	rootDir := cfg.RootDir

	// Remember the default listen addresses in case we need to apply the
	// default port to a user override.
	defaultListenJSONRPC := cfg.AppConfig.JSONRPCListenAddress

	// read in env config
	envCfg, err := LoadEnvConfig()
	if err != nil {
		return nil, false, fmt.Errorf("failed to load env config: %w", err)
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
		return nil, false, fmt.Errorf("failed to expand root directory \"%v\": %v", rootDir, err)
	}

	fmt.Printf("Root directory \"%v\"\n", rootDir)

	// make sure the root dir exists
	err = os.MkdirAll(rootDir, 0755)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create root directory \"%v\": %v", rootDir, err)
	}

	// 2. Read in the config file
	// read in config file and merge into default config
	var configFileExists bool
	fileCfg, err := LoadConfigFile(ConfigFilePath(rootDir))
	if err == nil {
		configFileExists = true
		// merge in config file
		err2 := cfg.Merge(fileCfg)
		if err2 != nil {
			return nil, false, fmt.Errorf("failed to merge config file: %w", err2)
		}
	} else if err != ErrConfigFileNotFound {
		return nil, false, fmt.Errorf("failed to load config file: %w", err)
	}

	// 3. Merge in the env config
	// merge in env config
	err = cfg.Merge(envCfg)
	if err != nil {
		return nil, false, fmt.Errorf("failed to merge env config: %w", err)
	}

	// 4. Merge in the flag config
	// merge in flag config
	err = cfg.Merge(flagCfg)
	if err != nil {
		return nil, false, fmt.Errorf("failed to merge flag config: %w", err)
	}

	cfg.RootDir = rootDir

	err = sanitizeCfgPaths(cfg)
	if err != nil {
		return nil, false, fmt.Errorf("failed to sanitize config paths: %w", err)
	}

	err = configureCerts(cfg)
	if err != nil {
		return nil, false, fmt.Errorf("failed to configure certs: %w", err)
	}

	if cfg.ChainConfig.Moniker == "" {
		cfg.ChainConfig.Moniker = defaultMoniker()
	}

	cfg.AppConfig.JSONRPCListenAddress = cleanListenAddr(cfg.AppConfig.JSONRPCListenAddress, defaultListenJSONRPC)

	// handling deprecation of configs:
	if cfg.AppConfig.DEPRECATED_RPCReqLimit != 0 {
		if cfg.AppConfig.RPCMaxReqSize != cmd.DefaultConfig().AppConfig.RPCMaxReqSize {
			return nil, false, fmt.Errorf("cannot set both deprecated rpc request limit and rpc max request size")
		}
		cfg.AppConfig.RPCMaxReqSize = cfg.AppConfig.DEPRECATED_RPCReqLimit
		fmt.Println("WARNING: app.rpc_req_limit is deprecated and will be removed in Kwil v0.10. use app.rpc_max_req_size instead")
	}

	if cfg.AppConfig.Snapshots.DEPRECATED_Enabled {
		if cfg.AppConfig.Snapshots.Enable {
			return nil, false, fmt.Errorf("cannot set both deprecated snapshots enabled and snapshots enable")
		}
		cfg.AppConfig.Snapshots.Enable = cfg.AppConfig.Snapshots.DEPRECATED_Enabled
		fmt.Println("WARNING: app.snapshots.enabled is deprecated and will be removed in Kwil v0.10. use app.snapshots.enable instead")
	}

	return cfg, configFileExists, nil
}

func configureCerts(cfg *config.KwildConfig) error {
	if cfg.AppConfig.TLSCertFile == "" {
		cfg.AppConfig.TLSCertFile = DefaultTLSCertFile
	}
	path, err := config.CleanPath(cfg.AppConfig.TLSCertFile, cfg.RootDir)
	if err != nil {
		return err
	}
	cfg.AppConfig.TLSCertFile = path

	if cfg.AppConfig.TLSKeyFile == "" {
		cfg.AppConfig.TLSKeyFile = defaultTLSKeyFile
	}
	path, err = config.CleanPath(cfg.AppConfig.TLSKeyFile, cfg.RootDir)
	if err != nil {
		return err
	}
	cfg.AppConfig.TLSKeyFile = path
	return nil
}

func sanitizeCfgPaths(cfg *config.KwildConfig) error {
	rootDir := cfg.RootDir

	path, err := config.CleanPath(cfg.AppConfig.PrivateKeyPath, rootDir)
	if err != nil {
		return fmt.Errorf("failed to expand private key path \"%v\": %v", cfg.AppConfig.PrivateKeyPath, err)
	}
	cfg.AppConfig.PrivateKeyPath = path
	fmt.Println("Private key path:", cfg.AppConfig.PrivateKeyPath)

	if cfg.AppConfig.GenesisState != "" {
		path, err := config.CleanPath(cfg.AppConfig.GenesisState, rootDir)
		if err != nil {
			return fmt.Errorf("failed to expand snapshot file path \"%v\": %v", cfg.AppConfig.GenesisState, err)
		}
		cfg.AppConfig.GenesisState = path
		fmt.Println("Snapshot file to initialize database from:", cfg.AppConfig.GenesisState)
	}

	return nil
}

// cleanListenAddr ensures that the provided listen includes both a host and
// port, using the host and port from defaultListen as needed.
func cleanListenAddr(listen, defaultListen string) string {
	defaultHost, defaultPort, _ := net.SplitHostPort(defaultListen) // empty if invalid default
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		var msg string
		addrErr := new(net.AddrError)
		if errors.As(err, &addrErr) {
			host = addrErr.Addr
			msg = addrErr.Err
		} else { // may be incorrect if host couldn't parse, but try
			host = listen
			msg = err.Error()
		}
		if strings.Contains(msg, "missing port") { // they really didn't export this :/
			host = strings.Trim(host, "[]")            // cut off brackets of an ipv6 addr
			return net.JoinHostPort(host, defaultPort) // no change if default had none
		}
		return listen // let the listener try
	}
	if host != "" && port != "" { // nothing missing
		return listen
	}
	if port == "" { // should be the "missing port" case above
		port = defaultPort // no change if default had none
	}
	if host == "" {
		host = defaultHost // no change if default had none
	}
	return net.JoinHostPort(host, port)
}

// LoadConfig reads a config.toml at the given path and returns a KwilConfig.
// If the file does not exist, it will return an ErrConfigFileNotFound error.
func LoadConfigFile(configPath string) (*config.KwildConfig, error) {
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
	var kwilCfg config.KwildConfig

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
				if t != reflect.TypeOf(config.Duration(time.Duration(5))) {
					return data, nil
				}

				// Convert it by parsing
				dur, err := time.ParseDuration(data.(string))
				if err != nil {
					return nil, err
				}

				return config.Duration(dur), nil
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
func LoadEnvConfig() (*config.KwildConfig, error) {
	// Manually bind environment variables to viper keys.
	for _, key := range viper.AllKeys() {
		// Replace dashes with underscores in the key to match the flag name.
		// This is required because there is inconsistency between our flag names
		// and the struct tags. The struct tags use underscores, but the flag names
		// use dashes. Viper uses the flag names to bind environment variables
		// and this conversion is required to map it to the struct fields correctly.
		bindKey := strings.ReplaceAll(key, "-", "_")
		envKey := "KWILD_" + strings.ToUpper(strings.ReplaceAll(bindKey, ".", "_"))
		viper.BindEnv(bindKey, envKey)
	}

	// TODO: try this
	// viper.SetEnvPrefix("KWILD")
	// viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")) // --app.output-paths => KWILD_APP_OUTPUT_PATHS
	// viper.AutomaticEnv()

	// var cfg KwildConfig, won't work because, viper won't be able to extract
	// the heirarchical keys from the config structure as fields like cfg.app set to nil.
	// It can only extract the first level keys [app, chain, log] in this case.
	// To remedy that, we use DefaultEmptyConfig with all the sub fields initialized.
	cfg := DefaultEmptyConfig()
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("decoding config: %v", err)
	}

	return cfg, nil
}

var ErrConfigFileNotFound = fmt.Errorf("config file not found")

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DefaultEmptyConfig returns a config with all fields set to their zero values.
// This is used by viper to extract all the heirarchical keys from the config
// structure.
func DefaultEmptyConfig() *config.KwildConfig {
	return &config.KwildConfig{
		AppConfig: &config.AppConfig{
			Extensions: make(map[string]map[string]string),
		},
		ChainConfig: &config.ChainConfig{
			P2P:       &config.P2PConfig{},
			RPC:       &config.ChainRPCConfig{},
			Mempool:   &config.MempoolConfig{},
			StateSync: &config.StateSyncConfig{},
			Consensus: &config.ConsensusConfig{},
		},
		Logging:         &config.Logging{},
		Instrumentation: &config.InstrumentationConfig{},
	}
}

// EmptyConfig returns a config with all fields set to their zero values (except
// no nil pointers for the sub-sections structs). This is useful for
// guaranteeing that all fields are set when merging.
func EmptyConfig() *config.KwildConfig {
	return &config.KwildConfig{
		AppConfig: &config.AppConfig{
			ExtensionEndpoints: []string{},
		},
		ChainConfig: &config.ChainConfig{
			P2P:       &config.P2PConfig{},
			RPC:       &config.ChainRPCConfig{},
			Mempool:   &config.MempoolConfig{},
			StateSync: &config.StateSyncConfig{},
			Consensus: &config.ConsensusConfig{},
		},
		Logging:         &config.Logging{},
		Instrumentation: &config.InstrumentationConfig{},
	}
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
		return nil, nil, false, fmt.Errorf("failed to load private key: %w", err)
	}

	priv, pub, err = newNodeKey()
	if err != nil {
		return nil, nil, false, err
	}

	return priv, pub, true, saveNodeKey(priv, keyPath)
}
