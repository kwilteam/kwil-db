// Package nodecfg provides functions to assist in the generation of new kwild
// node configurations. This is primarily intended for the kwil-admin commands
// and tests that required dynamic node configuration.
package nodecfg

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	// NOTE: do not use the types from these internal packages on nodecfg's
	// exported API.
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
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
	ChainID string
	// InitialHeight int64 // ?
	OutputDir       string
	JoinExpiry      int64
	WithoutGasCosts bool
	WithoutNonces   bool
}

type TestnetGenerateConfig struct {
	ChainID string
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
}

// GenerateTestnetConfig is like GenerateNodeConfig but it generates multiple
// configs for a network of nodes on a LAN.  See also TestnetGenerateConfig.
func GenerateTestnetConfig(genCfg *TestnetGenerateConfig) error {
	var err error
	genCfg.OutputDir, err = config.ExpandPath(genCfg.OutputDir)
	if err != nil {
		fmt.Println("Error while getting absolute path for output directory: ", err)
		return err
	}

	nNodes := genCfg.NValidators + genCfg.NNonValidators
	if nHosts := len(genCfg.Hostnames); nHosts > 0 && nHosts != nNodes {
		return fmt.Errorf(
			"testnet needs precisely %d hostnames (number of validators plus nonValidators) if --hostname parameter is used",
			nNodes,
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
		persistentPeers = persistentPeersString(genCfg, privateKeys)
	}

	// Overwrite default config
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

func persistentPeersString(genCfg *TestnetGenerateConfig, privKeys []cmtEd.PrivKey) string {
	persistentPeers := make([]string, genCfg.NValidators+genCfg.NNonValidators)
	for i := range persistentPeers {
		pubKey := privKeys[i].PubKey().(cmtEd.PubKey)
		hostPort := fmt.Sprintf("%s:%d", hostnameOrIP(genCfg, i), genCfg.P2pPort)
		persistentPeers[i] = cometbft.NodeIDAddressString(pubKey, hostPort)
	}
	return strings.Join(persistentPeers, ",")
}
