// Package nodecfg provides functions to assist in the generation of new kwild
// node configurations. This is primarily intended for the kwil-admin commands
// and tests that required dynamic node configuration.
package nodecfg

import (
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	// NOTE: do not use the types from these internal packages on nodecfg's
	// exported API.
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/core/types/chain"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"

	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
)

const (
	nodeDirPerm   = 0755
	chainIDPrefix = "kwil-chain-"

	abciDir       = config.ABCIDirName
	abciConfigDir = cometbft.ConfigDir
	abciDataDir   = cometbft.DataDir
)

type NodeGenerateConfig struct {
	ChainID       string
	BlockInterval time.Duration
	// InitialHeight int64 // ?
	OutputDir       string
	JoinExpiry      int64
	WithoutGasCosts bool
	WithoutNonces   bool

	ChainRPCURL           string
	EscrowAddress         string
	RequiredConfirmations int64
}

type TestnetGenerateConfig struct {
	ChainID       string
	BlockInterval time.Duration
	// InitialHeight           int64
	NValidators             int
	NNonValidators          int
	ConfigFile              string
	OutputDir               string
	NodeDirPrefix           string
	PopulatePersistentPeers bool
	HostnamePrefix          string
	HostnameSuffix          string
	StartingIPAddress       string
	Hostnames               []string
	P2pPort                 int
	JoinExpiry              int64
	WithoutGasCosts         bool
	WithoutNonces           bool
	ChainRPCURL             string
	EscrowAddress           string
	RequiredConfirmations   int64
}

// ConfigOpts is a struct to alter the generation of the node config.
type ConfigOpts struct {
	// UniquePorts is a flag to generate unique listening addresses
	// (gRPC, HTTP, Admin, P2P, RPC) for each node.
	// This is useful for testing multiple nodes on the same machine.
	// If it is used for generating a single config, it has no effect.
	UniquePorts bool
}

// GenerateNodeConfig is used to generate configuration required for running a
// kwil node. This includes the files: private_key, config.toml, genesis.json.
//
//   - The private key is generated if it does not exist.
//   - The genesis file is generated if it does not exist. A new genesis file
//     will include the node as a validator; existing genesis is not updated.
//   - The config.toml file is generated if it does not exist.
func GenerateNodeConfig(genCfg *NodeGenerateConfig) error {
	rootDir, err := config.ExpandPath(genCfg.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for output directory: %w", err)
	}

	cfg := config.DefaultConfig()
	cfg.RootDir = rootDir
	if genCfg.BlockInterval > 0 {
		cfg.ChainCfg.Consensus.TimeoutCommit = genCfg.BlockInterval
	}
	chainRoot := filepath.Join(rootDir, abciDir)
	// NOTE: not the fully re-rooted path since this may run in a container. The
	// caller can update PrivateKeyPath if desired.

	fullABCIConfigDir := filepath.Join(chainRoot, abciConfigDir)
	err = os.MkdirAll(fullABCIConfigDir, nodeDirPerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(chainRoot, abciDataDir), nodeDirPerm)
	if err != nil {
		return err
	}

	cfg.AppCfg.PrivateKeyPath = config.PrivateKeyFileName
	err = writeConfigFile(filepath.Join(rootDir, config.ConfigFileName), cfg)
	if err != nil {
		return err
	}

	// Load or generate private key.
	fullKeyPath := filepath.Join(rootDir, config.PrivateKeyFileName)
	_, pubKey, newKey, err := config.ReadOrCreatePrivateKeyFile(fullKeyPath, true)
	if err != nil {
		return fmt.Errorf("cannot read or create private key: %w", err)
	}
	if newKey {
		fmt.Printf("Generated new private key: %v\n", fullKeyPath)
	}

	// Create or update genesis config.
	genFile := filepath.Join(fullABCIConfigDir, cometbft.GenesisJSONName)

	_, err = os.Stat(genFile)
	if os.IsNotExist(err) {
		genesisCfg := config.NewGenesisWithValidator(pubKey)
		genCfg.applyGenesisParams(genesisCfg)
		return genesisCfg.SaveAs(genFile)
	}

	return err
}

func (genCfg *NodeGenerateConfig) applyGenesisParams(genesisCfg *config.GenesisConfig) {
	if genCfg.ChainID != "" {
		genesisCfg.ChainID = genCfg.ChainID
	}
	genesisCfg.ConsensusParams.Validator.JoinExpiry = genCfg.JoinExpiry
	genesisCfg.ConsensusParams.WithoutGasCosts = genCfg.WithoutGasCosts
	genesisCfg.ConsensusParams.WithoutNonces = genCfg.WithoutNonces

	genesisCfg.ConsensusParams.TokenBridge.Endpoint = genCfg.ChainRPCURL
	genesisCfg.ConsensusParams.TokenBridge.EscrowAddress = genCfg.EscrowAddress
	genesisCfg.ConsensusParams.TokenBridge.Code = chain.GOERLI
	genesisCfg.ConsensusParams.TokenBridge.RequiredConfirmations = genCfg.RequiredConfirmations

}

// GenerateTestnetConfig is like GenerateNodeConfig but it generates multiple
// configs for a network of nodes on a LAN.  See also TestnetGenerateConfig.
func GenerateTestnetConfig(genCfg *TestnetGenerateConfig, opts *ConfigOpts) error {
	if opts == nil {
		opts = &ConfigOpts{}
	}

	var err error
	genCfg.OutputDir, err = config.ExpandPath(genCfg.OutputDir)
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
	cfg := config.DefaultConfig()
	if genCfg.ConfigFile != "" {
		if err = cfg.ParseConfig(genCfg.ConfigFile); err != nil {
			return fmt.Errorf("failed to parse config file %s: %w", genCfg.ConfigFile, err)
		}
	}

	privateKeys := make([]cmtEd.PrivKey, nNodes)
	for i := range privateKeys {
		privateKeys[i] = cmtEd.GenPrivKey()

		nodeDirName := fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i)
		nodeDir := filepath.Join(genCfg.OutputDir, nodeDirName)
		chainRoot := filepath.Join(nodeDir, abciDir)

		err := os.MkdirAll(filepath.Join(chainRoot, abciConfigDir), nodeDirPerm)
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
		privKeyFile := filepath.Join(nodeDir, config.PrivateKeyFileName)
		err = os.WriteFile(privKeyFile, []byte(privKeyHex), 0644) // permissive for testnet only
		if err != nil {
			return fmt.Errorf("creating private key file: %w", err)
		}
	}

	genConfig := config.DefaultGenesisConfig()
	for i, pk := range privateKeys[:genCfg.NValidators] {
		genConfig.Validators = append(genConfig.Validators, &config.GenesisValidator{
			PubKey: pk.PubKey().Bytes(),
			Power:  1,
			Name:   fmt.Sprintf("validator-%d", i),
		})
	}
	genCfg.applyGenesisParams(genConfig)

	// write genesis file
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		chainRoot := filepath.Join(nodeDir, abciDir)
		genFile := cometbft.GenesisPath(chainRoot) // filepath.Join(nodeDir, abciDir, abciConfigDir, "genesis.json")
		err = genConfig.SaveAs(genFile)
		if err != nil {
			return fmt.Errorf("failed to write genesis file %v: %w", genFile, err)
		}
	}

	// Gather persistent peers addresses
	var persistentPeers string
	if genCfg.PopulatePersistentPeers {
		persistentPeers = persistentPeersString(genCfg, privateKeys, opts.UniquePorts)
	}

	// Overwrite default config
	if genCfg.BlockInterval > 0 {
		cfg.ChainCfg.Consensus.TimeoutCommit = genCfg.BlockInterval
	}
	cfg.ChainCfg.P2P.AddrBookStrict = false
	cfg.ChainCfg.P2P.AllowDuplicateIP = true
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		cfg.RootDir = nodeDir

		if genCfg.PopulatePersistentPeers {
			cfg.ChainCfg.P2P.PersistentPeers = persistentPeers
		}
		if i <= len(genCfg.Hostnames)-1 {
			cfg.AppCfg.Hostname = genCfg.Hostnames[i]
		}
		if i > 0 && opts.UniquePorts {
			err = addressSpecificConfig(cfg)
			if err != nil {
				return fmt.Errorf("failed to apply unique addresses: %w", err)
			}
		}

		cfg.AppCfg.PrivateKeyPath = config.PrivateKeyFileName // not abs/rooted because this might be run in a container
		writeConfigFile(filepath.Join(nodeDir, config.ConfigFileName), cfg)
	}

	fmt.Printf("Successfully initialized %d node directories: %s\n",
		genCfg.NValidators+genCfg.NNonValidators, genCfg.OutputDir)

	return nil
}

func (genCfg *TestnetGenerateConfig) applyGenesisParams(genesisCfg *config.GenesisConfig) {
	if genCfg.ChainID != "" {
		genesisCfg.ChainID = genCfg.ChainID
	}
	genesisCfg.ConsensusParams.Validator.JoinExpiry = genCfg.JoinExpiry
	genesisCfg.ConsensusParams.WithoutGasCosts = genCfg.WithoutGasCosts
	genesisCfg.ConsensusParams.WithoutNonces = genCfg.WithoutNonces
	genesisCfg.ConsensusParams.TokenBridge.Endpoint = genCfg.ChainRPCURL
	genesisCfg.ConsensusParams.TokenBridge.EscrowAddress = genCfg.EscrowAddress
	genesisCfg.ConsensusParams.TokenBridge.Code = chain.GOERLI
	genesisCfg.ConsensusParams.TokenBridge.RequiredConfirmations = genCfg.RequiredConfirmations

}

func hostnameOrIP(genCfg *TestnetGenerateConfig, i int) string {
	if len(genCfg.Hostnames) > 0 && i < len(genCfg.Hostnames) {
		return genCfg.Hostnames[i]
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
func persistentPeersString(genCfg *TestnetGenerateConfig, privKeys []cmtEd.PrivKey, decrementingPorts bool) string {
	persistentPeers := make([]string, genCfg.NValidators+genCfg.NNonValidators)
	for i := range persistentPeers {
		pubKey := privKeys[i].PubKey().(cmtEd.PubKey)

		port := genCfg.P2pPort
		if decrementingPorts {
			port -= i
		}

		hostPort := fmt.Sprintf("%s:%d", hostnameOrIP(genCfg, i), port)
		persistentPeers[i] = cometbft.NodeIDAddressString(pubKey, hostPort)
	}
	return strings.Join(persistentPeers, ",")
}

// applyUniqueAddresses applies unique addresses to the config.
// it will begin at the default port and increment by 1 for each node.
func addressSpecificConfig(c *config.KwildConfig) error {

	grpcAddr, err := incrementPort(c.AppCfg.GrpcListenAddress, 1)
	if err != nil {
		return err
	}
	c.AppCfg.GrpcListenAddress = grpcAddr

	httpAddr, err := incrementPort(c.AppCfg.HTTPListenAddress, 1)
	if err != nil {
		return err
	}
	c.AppCfg.HTTPListenAddress = httpAddr

	adminAddr, err := incrementPort(c.AppCfg.AdminListenAddress, 1)
	if err != nil {
		return err
	}
	c.AppCfg.AdminListenAddress = adminAddr

	p2pAddr, err := incrementPort(c.ChainCfg.P2P.ListenAddress, -1) // decrement since default rpc is 1 higher than p2p, so p2p needs to be 1 lower
	if err != nil {
		return err
	}
	c.ChainCfg.P2P.ListenAddress = p2pAddr

	rpcAddr, err := incrementPort(c.ChainCfg.RPC.ListenAddress, 1)
	if err != nil {
		return err
	}
	c.ChainCfg.RPC.ListenAddress = rpcAddr

	return nil
}

// incrementPort increments the port in the URL by the given amount.
func incrementPort(incoming string, amt int) (string, error) {
	res, err := url.Parse(incoming)
	if err != nil {
		return "", err
	}

	// Split the URL into two parts: host (and possibly scheme) and port
	host, portStr, err := net.SplitHostPort(res.Host)
	if err != nil {
		// Handle the case where there is no scheme (like "localhost:50051")
		if strings.Contains(err.Error(), "missing port in address") {
			parts := strings.Split(incoming, ":")
			if len(parts) != 2 {
				return "", fmt.Errorf("invalid URL format")
			}
			host, portStr = parts[0], parts[1]
		} else {
			return "", err
		}
	}

	// Convert the port to an integer and increment it
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", err
	}
	port += amt

	// Reconstruct and return the new URL
	newUrl := net.JoinHostPort(host, strconv.Itoa(port))

	if res.Scheme != "localhost" {
		newUrl = res.Scheme + "://" + newUrl
	}

	return newUrl, nil
}
