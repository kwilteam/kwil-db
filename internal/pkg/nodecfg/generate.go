package nodecfg

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/p2p"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/spf13/viper"
)

const (
	nodeDirPerm   = 0755
	chainIDPrefix = "kwil-chain-"
)

type NodeGenerateConfig struct {
	InitialHeight int64
	HomeDir       string
}

type TestnetGenerateConfig struct {
	NValidators             int
	NNonValidators          int
	InitialHeight           int64
	ConfigFile              string
	OutputDir               string
	NodeDirPrefix           string
	PopulatePersistentPeers bool
	HostnamePrefix          string
	HostnameSuffix          string
	StartingIPAddress       string
	Hostnames               []string
	P2pPort                 int
}

func GenerateNodeConfig(genCfg *NodeGenerateConfig) error {
	cfg := config.DefaultConfig()
	cfg.RootDir = genCfg.HomeDir

	err := initFilesWithConfig(cfg, 0)
	if err != nil {
		return err
	}

	cfg.ChainCfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	WriteConfigFile(filepath.Join(genCfg.HomeDir, "abci/config", "config.toml"), cfg)
	return nil
}

func GenerateTestnetConfig(genCfg *TestnetGenerateConfig) error {
	if len(genCfg.Hostnames) > 0 && len(genCfg.Hostnames) != (genCfg.NValidators+genCfg.NNonValidators) {
		return fmt.Errorf(
			"testnet needs precisely %d hostnames (number of validators plus nonValidators) if --hostname parameter is used",
			genCfg.NValidators+genCfg.NNonValidators,
		)
	}

	privateKeys := make([]ed25519.PrivKey, genCfg.NValidators+genCfg.NNonValidators)
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		privateKeys[i] = ed25519.GenPrivKey()
	}

	cfg := config.DefaultConfig()

	// overwrite default config if set and valid
	if genCfg.ConfigFile != "" {
		viper.SetConfigFile(genCfg.ConfigFile)
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
		if err := viper.Unmarshal(cfg); err != nil {
			return err
		}
		if err := cfg.ChainCfg.ValidateBasic(); err != nil {
			return err
		}
	}

	genVals := make([]types.GenesisValidator, genCfg.NValidators)

	for i := 0; i < genCfg.NValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i)
		nodeDir := filepath.Join(genCfg.OutputDir, nodeDirName)
		cfg.RootDir = nodeDir
		cfg.ChainCfg.SetRoot(filepath.Join(nodeDir, "abci"))

		err := os.MkdirAll(filepath.Join(nodeDir, "abci", "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}

		err = os.MkdirAll(filepath.Join(nodeDir, "abci", "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}
		priv := privateKeys[i]
		cfg.AppCfg.PrivateKey = hex.EncodeToString(priv[:])
		pub := priv.PubKey().(ed25519.PubKey)
		addr := pub.Address()

		cfg.ChainCfg.P2P.AddrBookStrict = false
		cfg.ChainCfg.P2P.AllowDuplicateIP = true

		genVals[i] = types.GenesisValidator{
			Address: addr,
			PubKey:  pub,
			Power:   1,
			Name:    fmt.Sprintf("validator-%d", i),
		}
	}

	for i := 0; i < genCfg.NNonValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i+genCfg.NValidators)
		nodeDir := filepath.Join(genCfg.OutputDir, nodeDirName)
		cfg.RootDir = nodeDir

		err := os.MkdirAll(filepath.Join(nodeDir, "abci", "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(nodeDir)
			return err
		}

		err = os.MkdirAll(filepath.Join(nodeDir, "abci", "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(nodeDir)
			return err
		}
	}

	genDoc := &types.GenesisDoc{
		ChainID:         chainIDPrefix + cmtrand.Str(6),
		ConsensusParams: types.DefaultConsensusParams(),
		GenesisTime:     cmttime.Now(),
		InitialHeight:   genCfg.InitialHeight,
		Validators:      genVals,
	}

	// write genesis file
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		if err := genDoc.SaveAs(filepath.Join(nodeDir, "abci", "config", "genesis.json")); err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}
	}

	// Gather persistent peers addresses
	var (
		persistentPeers string
		err             error
	)

	if genCfg.PopulatePersistentPeers {
		persistentPeers, err = persistentPeersString(cfg, genCfg, privateKeys)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}
	}

	// Overwrite default config
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		cfg.RootDir = nodeDir
		cfg.ChainCfg.SetRoot(nodeDir)
		cfg.ChainCfg.P2P.AddrBookStrict = false
		cfg.ChainCfg.P2P.AllowDuplicateIP = true
		if genCfg.PopulatePersistentPeers {
			cfg.ChainCfg.P2P.PersistentPeers = persistentPeers
		}
		WriteConfigFile(filepath.Join(nodeDir, "abci", "config", "config.toml"), cfg)
	}

	fmt.Printf("Successfully initialized %d node directories\n", genCfg.NValidators)

	return nil
}

// It generates private keys, which we should not leave up to Comet
func initFilesWithConfig(cfg *config.KwildConfig, nodeIdx int) error {
	cfg.ChainCfg.SetRoot(filepath.Join(cfg.RootDir, "abci"))
	err := os.MkdirAll(filepath.Join(cfg.RootDir, "abci", "config"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(cfg.RootDir)
		return err
	}

	err = os.MkdirAll(filepath.Join(cfg.RootDir, "abci", "data"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(cfg.RootDir)
		return err
	}
	cfg.ChainCfg.P2P.AddrBookStrict = false
	cfg.ChainCfg.P2P.AllowDuplicateIP = true

	priv := ed25519.GenPrivKey()
	pub := priv.PubKey().(ed25519.PubKey)
	addr := pub.Address()
	cfg.AppCfg.PrivateKey = hex.EncodeToString(priv[:])
	NodeID := p2p.ID(hex.EncodeToString(addr))
	fmt.Println("NodeID for node-", nodeIdx, NodeID)

	// genesis file
	genFile := cfg.ChainCfg.GenesisFile()
	if cmtos.FileExists(genFile) {
		fmt.Printf("Found genesis file %v\n", genFile)
	} else {
		genDoc := types.GenesisDoc{
			ChainID:         chainIDPrefix + cmtrand.Str(6),
			GenesisTime:     cmttime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		}
		genDoc.Validators = []types.GenesisValidator{{
			Address: addr,
			PubKey:  pub,
			Power:   1,
			Name:    fmt.Sprintf("validator-%d", nodeIdx),
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}

		fmt.Printf("Generated genesis file %v\n", genFile)
	}
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
		fmt.Printf("%v: non ipv4 address\n", genCfg.StartingIPAddress)
		os.Exit(1)
	}

	for j := 0; j < i; j++ {
		ip[3]++
	}
	return ip.String()
}

func persistentPeersString(cfg *config.KwildConfig, genCfg *TestnetGenerateConfig, PrivKeys []ed25519.PrivKey) (string, error) {
	persistentPeers := make([]string, genCfg.NValidators+genCfg.NNonValidators)
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		cfg.RootDir = nodeDir
		cfg.ChainCfg.SetRoot(filepath.Join(nodeDir, "abci"))
		pub := PrivKeys[i].PubKey().(ed25519.PubKey)
		addr := pub.Address()
		nodeID := p2p.ID(hex.EncodeToString(addr))
		persistentPeers[i] = p2p.IDAddressString(nodeID, fmt.Sprintf("%s:%d", hostnameOrIP(genCfg, i), genCfg.P2pPort))
	}
	return strings.Join(persistentPeers, ","), nil
}
