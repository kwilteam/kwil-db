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

	"github.com/kwilteam/kwil-db/internal/app/kwild"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft"

	"github.com/cometbft/cometbft/crypto/ed25519"
)

const (
	nodeDirPerm   = 0755
	chainIDPrefix = "kwil-chain-"

	abciDir       = kwild.ABCIDirName
	abciConfigDir = cometbft.ConfigDir
	abciDataDir   = cometbft.DataDir
)

type NodeGenerateConfig struct {
	// InitialHeight int64 // ?
	OutputDir       string
	JoinExpiry      int64
	WithoutGasCosts bool
	WithoutNonces   bool
}

type TestnetGenerateConfig struct {
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

/*
 GenerateNodeConfig is used to generate configuration required for running a kwil node.
	- private_key, config.toml, genesis.json

 The private key is generated if it does not exist.
 The genesis file is generated if it does not exist (or) updated if a new private key is generated.
 The config.toml file is generated if it does not exist.
*/

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

	err = os.MkdirAll(filepath.Join(chainRoot, abciConfigDir), nodeDirPerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(chainRoot, abciDataDir), nodeDirPerm)
	if err != nil {
		return err
	}

	cfg.AppCfg.PrivateKeyPath = kwild.PrivateKeyFileName

	genParams := &config.GenesisParams{
		JoinExpiry:      genCfg.JoinExpiry,
		WithoutGasCosts: genCfg.WithoutGasCosts,
		WithoutNonces:   genCfg.WithoutNonces,
		ChainIDPrefix:   chainIDPrefix,
	}
	_, _, _, _, err = config.LoadGenesisAndPrivateKey(true,
		filepath.Join(cfg.RootDir, kwild.PrivateKeyFileName),
		chainRoot, genParams)
	if err != nil {
		return err
	}

	// cometbft's gRPC server -- we won't use, right?
	// cfg.ChainCfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"

	writeConfigFile(filepath.Join(rootDir, kwild.ConfigFileName), cfg)

	fmt.Println("Successfully initialized node directory: ", rootDir)
	return nil
}

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

	privateKeys := make([]ed25519.PrivKey, nNodes)
	for i := range privateKeys {
		privateKeys[i] = ed25519.GenPrivKey()

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
		privKeyFile := filepath.Join(nodeDir, kwild.PrivateKeyFileName)
		err = os.WriteFile(privKeyFile, []byte(privKeyHex), 0644) // permissive for testnet only
		if err != nil {
			return fmt.Errorf("creating private key file: %w", err)
		}
	}

	validatorPkeys := privateKeys[:genCfg.NValidators]
	genParams := &config.GenesisParams{
		JoinExpiry:      genCfg.JoinExpiry,
		WithoutGasCosts: genCfg.WithoutGasCosts,
		WithoutNonces:   genCfg.WithoutNonces,
		ChainIDPrefix:   chainIDPrefix,
	}
	genConfig := config.GenerateGenesisConfig(validatorPkeys, genParams)

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
		cfg.AppCfg.PrivateKeyPath = kwild.PrivateKeyFileName // not abs/rooted because this might be run in a container
		writeConfigFile(filepath.Join(nodeDir, kwild.ConfigFileName), cfg)
	}

	fmt.Printf("Successfully initialized %d node directories: %s\n",
		genCfg.NValidators+genCfg.NNonValidators, genCfg.OutputDir)

	return nil
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

func persistentPeersString(genCfg *TestnetGenerateConfig, privKeys []ed25519.PrivKey) string {
	persistentPeers := make([]string, genCfg.NValidators+genCfg.NNonValidators)
	for i := range persistentPeers {
		pubKey := privKeys[i].PubKey().(ed25519.PubKey)
		hostPort := fmt.Sprintf("%s:%d", hostnameOrIP(genCfg, i), genCfg.P2pPort)
		persistentPeers[i] = cometbft.NodeIDAddressString(pubKey, hostPort)
	}
	return strings.Join(persistentPeers, ",")
}
