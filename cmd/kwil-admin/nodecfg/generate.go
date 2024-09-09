// Package nodecfg provides functions to assist in the generation of new kwild
// node configurations. This is primarily intended for the kwil-admin commands
// and tests that required dynamic node configuration.
package nodecfg

import (
	"encoding/hex"
	"fmt"
	"maps"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	// NOTE: do not use the types from these internal packages on nodecfg's
	// exported API.
	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/kwilteam/kwil-db/cmd"
	kwildcfg "github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common/chain"
	config "github.com/kwilteam/kwil-db/common/config"
	coreUrl "github.com/kwilteam/kwil-db/core/utils/url"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
)

const (
	nodeDirPerm   = 0755
	chainIDPrefix = "kwil-chain-"

	abciDir     = kwildcfg.ABCIDirName
	abciDataDir = cometbft.DataDir
)

var (
	genesisValidatorGas, _ = big.NewInt(0).SetString("10000000000000000000000", 10)
)

type NodeGenerateConfig struct {
	ChainID       string
	BlockInterval time.Duration
	// InitialHeight int64 // ?
	OutputDir       string
	JoinExpiry      int64
	WithoutGasCosts bool
	Allocs          map[string]*big.Int
	VoteExpiry      int64
	Extensions      map[string]map[string]string
	Forks           map[string]*uint64
}

type TestnetGenerateConfig struct {
	ChainID       string
	BlockInterval time.Duration
	// InitialHeight           int64
	AdminAddress string
	AdminNoTLS   bool

	NValidators             int
	NNonValidators          int
	ConfigFile              string
	OutputDir               string
	NodeDirPrefix           string
	PopulatePersistentPeers bool
	HostnamePrefix          string
	HostnameSuffix          string
	StartingIPAddress       string
	DnsNamePrefix           string
	Hostnames               []string
	P2pPort                 int
	JoinExpiry              int64
	WithoutGasCosts         bool
	VoteExpiry              int64
	Allocs                  map[string]*big.Int
	FundNonValidators       bool
	Extensions              []map[string]map[string]string // for each node
	SnapshotsEnabled        bool
	MaxSnapshots            uint64
	SnapshotHeights         uint64
	Forks                   map[string]*uint64
	PrivateMode             bool
}

// ConfigOpts is a struct to alter the generation of the node config.
type ConfigOpts struct {
	// UniquePorts is a flag to generate unique listening addresses
	// (JSON-RPC, HTTP, Admin, P2P, node RPC) for each node.
	// This is useful for testing multiple nodes on the same machine.
	// If it is used for generating a single config, it has no effect.
	UniquePorts bool

	// NoGenesis is a flag to not generate a genesis file.
	// This is useful if you are generating a node config for
	// a network that already has a genesis file.
	// If used with creating a testnet, it will result in an error.
	NoGenesis bool

	// DnsHost is a flag to use DNS hostname as host in the config
	// instead of ip. It will be used together with DnsNamePrefix to generate
	// hostnames.
	// This is useful for testing nodes inside docker containers.
	DnsHost bool
}

// GenerateNodeConfig is used to generate configuration required for running a
// kwil node. This includes the files: private_key, config.toml, genesis.json.
//
//   - The private key is generated if it does not exist.
//   - The genesis file is generated if it does not exist. A new genesis file
//     will include the node as a validator; existing genesis is not updated.
//   - The config.toml file is generated if it does not exist.
func GenerateNodeConfig(genCfg *NodeGenerateConfig) error {
	rootDir, err := kwildcfg.ExpandPath(genCfg.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for output directory: %w", err)
	}

	cfg := cmd.DefaultConfig()
	cfg.RootDir = rootDir
	if genCfg.BlockInterval > 0 {
		cfg.ChainConfig.Consensus.TimeoutCommit = config.Duration(genCfg.BlockInterval)
	}

	if genCfg.Extensions != nil {
		cfg.AppConfig.Extensions = genCfg.Extensions
	} else {
		cfg.AppConfig.Extensions = make(map[string]map[string]string)
	}

	pub, err := GenerateNodeFiles(rootDir, cfg, false)
	if err != nil {
		return err
	}

	// Create or update genesis config.
	genFile := filepath.Join(rootDir, cometbft.GenesisJSONName)

	_, err = os.Stat(genFile)
	if os.IsNotExist(err) {
		genesisCfg := chain.NewGenesisWithValidator(pub)
		genCfg.ApplyGenesisParams(genesisCfg)
		return genesisCfg.SaveAs(genFile)

	}

	return err
}

// GenerateNodeFiles will generate all generic node files that are not
// dependent on the network configuration (e.g. genesis.json).
// It can optionally be given a config file to merge with the default config.
func GenerateNodeFiles(outputDir string, originalCfg *config.KwildConfig, silence bool) (publicKey []byte, err error) {
	cfg := cmd.DefaultConfig()
	if originalCfg != nil {
		err := cfg.Merge(originalCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to merge config file: %w", err)
		}
	}

	rootDir, err := kwildcfg.ExpandPath(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for output directory: %w", err)
	}
	cfg.RootDir = rootDir

	cometbftRoot := filepath.Join(outputDir, abciDir)

	err = os.MkdirAll(cometbftRoot, nodeDirPerm)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(cometbftRoot, abciDataDir), nodeDirPerm)
	if err != nil {
		return nil, err
	}

	cfg.AppConfig.PrivateKeyPath = cmd.DefaultConfig().AppConfig.PrivateKeyPath
	err = WriteConfigFile(filepath.Join(rootDir, kwildcfg.ConfigFileName), cfg)
	if err != nil {
		return nil, err
	}

	// Load or generate private key.
	fullKeyPath := filepath.Join(rootDir, cmd.DefaultConfig().AppConfig.PrivateKeyPath)
	_, pubKey, newKey, err := kwildcfg.ReadOrCreatePrivateKeyFile(fullKeyPath, true)
	if err != nil {
		return nil, fmt.Errorf("cannot read or create private key: %w", err)
	}
	if newKey && !silence {
		fmt.Printf("Generated new private key: %v\n", fullKeyPath)
	}

	return pubKey, nil
}

func (genCfg *NodeGenerateConfig) ApplyGenesisParams(genesisCfg *chain.GenesisConfig) {
	if genCfg.ChainID != "" {
		genesisCfg.ChainID = genCfg.ChainID
	}
	genesisCfg.ConsensusParams.Validator.JoinExpiry = genCfg.JoinExpiry
	genesisCfg.ConsensusParams.WithoutGasCosts = genCfg.WithoutGasCosts
	genesisCfg.ConsensusParams.Votes.VoteExpiry = genCfg.VoteExpiry

	numAllocs := len(genCfg.Allocs)
	if !genCfg.WithoutGasCosts { // when gas is enabled, give genesis validators some funds
		numAllocs += len(genesisCfg.Validators)
	}
	if numAllocs > 0 {
		genesisCfg.Alloc = make(chain.GenesisAlloc, len(genCfg.Allocs))
		for acct, bal := range genCfg.Allocs {
			genesisCfg.Alloc[acct] = bal
		}
		for _, vi := range genesisCfg.Validators {
			genesisCfg.Alloc[vi.PubKey.String()] = genesisValidatorGas
		}
	}
	genesisCfg.ForkHeights = maps.Clone(genCfg.Forks)
}

// GenerateTestnetConfig is like GenerateNodeConfig but it generates multiple
// configs for a network of nodes on a LAN.  See also TestnetGenerateConfig.
func GenerateTestnetConfig(genCfg *TestnetGenerateConfig, opts *ConfigOpts) error {
	if opts == nil {
		opts = &ConfigOpts{}
	}
	if opts.NoGenesis {
		return fmt.Errorf("cannot use NoGenesis opt with testnet")
	}

	var err error
	genCfg.OutputDir, err = kwildcfg.ExpandPath(genCfg.OutputDir)
	if err != nil {
		fmt.Println("Error while getting absolute path for output directory: ", err)
		return err
	}

	nNodes := genCfg.NValidators + genCfg.NNonValidators
	if nHosts := len(genCfg.Hostnames); nHosts > 0 && nHosts != nNodes {
		return fmt.Errorf(
			"testnet needs precisely %d hostnames (for the %d validators and %d non-validators) if --hostnames parameter is used",
			nNodes, genCfg.NValidators, genCfg.NNonValidators,
		)
	}

	// overwrite default config if set and valid
	cfg := cmd.DefaultConfig()
	if genCfg.ConfigFile != "" {
		configFile, err := kwildcfg.LoadConfigFile(genCfg.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load config file %s: %w", genCfg.ConfigFile, err)
		}

		err = cfg.Merge(configFile)
		if err != nil {
			return fmt.Errorf("failed to merge config file %s: %w", genCfg.ConfigFile, err)
		}
	}

	if genCfg.AdminAddress != "" {
		cfg.AppConfig.AdminListenAddress = genCfg.AdminAddress
	}
	cfg.AppConfig.NoTLS = genCfg.AdminNoTLS

	privateKeys := make([]cmtEd.PrivKey, nNodes)
	for i := range privateKeys {
		privateKeys[i] = cmtEd.GenPrivKey()

		nodeDirName := fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i)
		nodeDir := filepath.Join(genCfg.OutputDir, nodeDirName)
		chainRoot := filepath.Join(nodeDir, abciDir)

		err := os.MkdirAll(chainRoot, nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}

		err = os.MkdirAll(filepath.Join(chainRoot, abciDataDir), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}

		privKeyHex := hex.EncodeToString(privateKeys[i][:])
		privKeyFile := filepath.Join(nodeDir, cmd.DefaultConfig().AppConfig.PrivateKeyPath)
		err = os.WriteFile(privKeyFile, []byte(privKeyHex), 0644) // permissive for testnet only
		if err != nil {
			return fmt.Errorf("creating private key file: %w", err)
		}
	}

	genConfig := chain.DefaultGenesisConfig()
	for i, pk := range privateKeys[:genCfg.NValidators] {
		genConfig.Validators = append(genConfig.Validators, &chain.GenesisValidator{
			PubKey: pk.PubKey().Bytes(),
			Power:  1,
			Name:   fmt.Sprintf("validator-%d", i),
		})
	}
	genCfg.applyGenesisParams(genConfig)
	if genCfg.FundNonValidators {
		for _, pk := range privateKeys[genCfg.NValidators:] {
			genConfig.Alloc[hex.EncodeToString(pk.PubKey().Bytes())] = genesisValidatorGas
		}
	}

	// write genesis file
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		genFile := filepath.Join(nodeDir, cometbft.GenesisJSONName)
		err = genConfig.SaveAs(genFile)
		if err != nil {
			return fmt.Errorf("failed to write genesis file %v: %w", genFile, err)
		}
	}

	// Gather persistent peers addresses
	var persistentPeers string
	if genCfg.PopulatePersistentPeers {
		persistentPeers = persistentPeersString(genCfg, privateKeys, opts.UniquePorts, opts.DnsHost)
	}

	// Overwrite default config
	if genCfg.BlockInterval > 0 {
		cfg.ChainConfig.Consensus.TimeoutCommit = config.Duration(genCfg.BlockInterval)
	}
	cfg.ChainConfig.P2P.AddrBookStrict = false
	cfg.ChainConfig.P2P.AllowDuplicateIP = true
	// private mode
	cfg.ChainConfig.P2P.PrivateMode = genCfg.PrivateMode

	if genCfg.SnapshotsEnabled {
		cfg.AppConfig.Snapshots.Enabled = true

		if genCfg.MaxSnapshots != 0 {
			cfg.AppConfig.Snapshots.MaxSnapshots = genCfg.MaxSnapshots
		}

		if genCfg.SnapshotHeights != 0 {
			cfg.AppConfig.Snapshots.RecurringHeight = genCfg.SnapshotHeights
		}
	}

	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		cfg.RootDir = nodeDir

		if genCfg.PopulatePersistentPeers {
			cfg.ChainConfig.P2P.PersistentPeers = persistentPeers
		}
		if i <= len(genCfg.Hostnames)-1 {
			cfg.AppConfig.Hostname = genCfg.Hostnames[i]
		}
		if opts.UniquePorts {
			err = uniqueAdminAddress(cfg)
			if err != nil {
				return fmt.Errorf("failed to apply unique admin address: %w", err)
			}

			if i > 0 {
				err = addressSpecificConfig(cfg)
				if err != nil {
					return fmt.Errorf("failed to apply unique addresses: %w", err)
				}
			}
		}

		cfg.AppConfig.PrivateKeyPath = cmd.DefaultConfig().AppConfig.PrivateKeyPath // not abs/rooted because this might be run in a container

		// extension config
		if i < genCfg.NValidators {
			if genCfg.Extensions != nil {
				if len(genCfg.Extensions) != genCfg.NValidators {
					return fmt.Errorf("extensions must be nil or have the same length as the number of validators")
				}
				cfg.AppConfig.Extensions = genCfg.Extensions[i]
			} else {
				cfg.AppConfig.Extensions = make(map[string]map[string]string)
			}
		}

		WriteConfigFile(filepath.Join(nodeDir, kwildcfg.ConfigFileName), cfg)
	}

	fmt.Printf("Successfully initialized %d node directories: %s\n",
		genCfg.NValidators+genCfg.NNonValidators, genCfg.OutputDir)

	return nil
}

func (genCfg *TestnetGenerateConfig) applyGenesisParams(genesisCfg *chain.GenesisConfig) {
	if genCfg.ChainID != "" {
		genesisCfg.ChainID = genCfg.ChainID
	}
	genesisCfg.ConsensusParams.Validator.JoinExpiry = genCfg.JoinExpiry
	genesisCfg.ConsensusParams.Votes.VoteExpiry = genCfg.VoteExpiry
	genesisCfg.ConsensusParams.WithoutGasCosts = genCfg.WithoutGasCosts
	numAllocs := len(genCfg.Allocs)
	if !genCfg.WithoutGasCosts { // when gas is enabled, give genesis validators some funds
		numAllocs += len(genesisCfg.Validators)
	}
	if numAllocs > 0 {
		genesisCfg.Alloc = make(chain.GenesisAlloc, len(genCfg.Allocs))
		for acct, bal := range genCfg.Allocs {
			genesisCfg.Alloc[acct] = bal
		}
		for _, vi := range genesisCfg.Validators {
			genesisCfg.Alloc[vi.PubKey.String()] = genesisValidatorGas
		}
	}
	genesisCfg.ForkHeights = maps.Clone(genCfg.Forks)
}

func hostnameOrIP(genCfg *TestnetGenerateConfig, i int, useDnsHost bool) string {
	if len(genCfg.Hostnames) > 0 && i < len(genCfg.Hostnames) {
		return genCfg.Hostnames[i]
	}
	if useDnsHost {
		return fmt.Sprintf("%s%d", genCfg.DnsNamePrefix, i)
	}
	if genCfg.StartingIPAddress == "" {
		return fmt.Sprintf("%s%d%s", genCfg.HostnamePrefix, i, genCfg.HostnameSuffix)
	}
	ip := net.ParseIP(genCfg.StartingIPAddress)
	ip = ip.To4()
	if ip == nil {
		panic(fmt.Sprintf("%v: non ipv4 address\n", genCfg.StartingIPAddress))
	}

	ip[3] += byte(i)
	return ip.String()
}

// persistentPeersString returns a comma-separated list of persistent peers
// if decrementingPorts is true, it will begin at the default port and
// decrement by 1 for each node.
func persistentPeersString(genCfg *TestnetGenerateConfig, privKeys []cmtEd.PrivKey, decrementingPorts bool,
	useDnsHost bool) string {
	persistentPeers := make([]string, genCfg.NValidators+genCfg.NNonValidators)
	for i := range persistentPeers {
		pubKey := privKeys[i].PubKey().(cmtEd.PubKey)

		port := genCfg.P2pPort
		if decrementingPorts {
			port -= i
		}

		hostPort := fmt.Sprintf("%s:%d", hostnameOrIP(genCfg, i, useDnsHost), port)
		persistentPeers[i] = cometbft.NodeIDAddressString(pubKey, hostPort)
	}
	return strings.Join(persistentPeers, ",")
}

// applyUniqueAddresses applies unique addresses to the config.
// it will begin at the default port and increment by 1 for each node.
// it does not apply to the admin address. This does NOT change the
// admin service address.
func addressSpecificConfig(c *config.KwildConfig) error {

	jsonrpcAddr, err := incrementPort(c.AppConfig.JSONRPCListenAddress, -1) // decrement to avoid collision with admin rpc at 8485
	if err != nil {
		return err
	}
	c.AppConfig.JSONRPCListenAddress = jsonrpcAddr

	p2pAddr, err := incrementPort(c.ChainConfig.P2P.ListenAddress, -1) // decrement since default rpc is 1 higher than p2p, so p2p needs to be 1 lower
	if err != nil {
		return err
	}
	c.ChainConfig.P2P.ListenAddress = p2pAddr

	rpcAddr, err := incrementPort(c.ChainConfig.RPC.ListenAddress, 1)
	if err != nil {
		return err
	}
	c.ChainConfig.RPC.ListenAddress = rpcAddr

	return nil
}

// uniqueAdminAddress applies a unique address to the config. This only works
// for host:port or unix socket paths, not URLs.
func uniqueAdminAddress(cfg *config.KwildConfig) error {
	addr := cfg.AppConfig.AdminListenAddress

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = addr   // this may be a unix path, checked below
			port = "8485" // for sequential addresses, use 8485 to make collision with 8484 unlikely
		} else if strings.Contains(err.Error(), "too many colons in address") {
			u, err := url.Parse(addr)
			if err != nil {
				return fmt.Errorf("unknown admin service address: %w", err)
			}
			host, port = u.Hostname(), u.Port()
		} else {
			return fmt.Errorf("unknown admin service address: %w", err)
		}
	}

	if isUNIX := strings.HasPrefix(host, "/"); !isUNIX {
		addr = net.JoinHostPort(host, port)
		addr, err = incrementPort(addr, 1)
		if err != nil {
			return err
		}

		cfg.AppConfig.AdminListenAddress = addr
		return nil
	}

	extension := filepath.Ext(host)
	fileWithoutExt := strings.TrimSuffix(host, extension)

	// see if the file already has a number appended to it
	numberToUse := 0

	nums := strings.Split(fileWithoutExt, "_")
	if len(nums) > 1 {
		last := nums[len(nums)-1]
		// if the last part is a number, remove it
		num, err := strconv.Atoi(last)
		if err == nil {
			fileWithoutExt = strings.TrimSuffix(fileWithoutExt, "_"+last)
			numberToUse = num + 1
		}
	}

	cfg.AppConfig.AdminListenAddress = fileWithoutExt + "_" + strconv.Itoa(numberToUse) + extension

	return nil
}

// incrementPort increments the port in the URL by the given amount.
// if the url is a UNIX socket, it will append the amount to the path.
func incrementPort(incoming string, amt int) (string, error) {
	schemaExists, err := coreUrl.HasScheme(incoming)
	if err != nil {
		return "", err
	}

	if !schemaExists {
		incoming = "tcp://" + incoming
	}

	res, err := url.Parse(incoming)
	if err != nil {
		return "", err
	}

	// Split the URL into two parts: host (and possibly scheme) and port
	host, portStr, err := net.SplitHostPort(res.Host)
	if err != nil {
		// err will occur if there is no scheme, complaining (incorrectly) that
		// there is no port. If res.Opaque is not empty, then
		// res.Scheme is actually the host, and
		// res.Opaque is the port.
		if res.Opaque == "" {
			return "", err
		}
		host = res.Scheme
		portStr = res.Opaque
	}

	// Convert the port to an integer and increment it
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", err
	}
	port += amt

	// Reconstruct and return the new URL
	newUrl := net.JoinHostPort(host, strconv.Itoa(port))

	if schemaExists {
		newUrl = res.Scheme + "://" + newUrl
	}

	return newUrl, nil
}
