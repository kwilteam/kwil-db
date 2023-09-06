package nodecfg

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"

	cmtos "github.com/cometbft/cometbft/libs/os"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"

	"github.com/spf13/viper"
)

const (
	nodeDirPerm   = 0755
	chainIDPrefix = "kwil-chain-"
	PrivKeyFile   = "private_key.txt"
)

type NodeGenerateConfig struct {
	InitialHeight int64
	OutputDir     string
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

	rootDir, err := config.ExpandPath(genCfg.OutputDir)
	if err != nil {
		fmt.Println("Error while getting absolute path for output directory: ", err)
		return err
	}

	cfg.RootDir = rootDir
	err = initFilesWithConfig(cfg, 0)
	if err != nil {
		return err
	}

	cfg.ChainCfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"

	writeConfigFile(filepath.Join(rootDir, "config.toml"), cfg)

	fmt.Println("Successfully initialized node directory: ", rootDir)
	return nil
}

func GenerateTestnetConfig(genCfg *TestnetGenerateConfig) ([]ed25519.PrivKey, error) {
	rootDir, err := config.ExpandPath(genCfg.OutputDir)
	if err != nil {
		fmt.Println("Error while getting absolute path for output directory: ", err)
		return nil, err
	}
	genCfg.OutputDir = rootDir

	if len(genCfg.Hostnames) > 0 && len(genCfg.Hostnames) != (genCfg.NValidators+genCfg.NNonValidators) {
		return nil, fmt.Errorf(
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
			return nil, err
		}
		if err := viper.Unmarshal(cfg); err != nil {
			return nil, err
		}
		if err := cfg.ChainCfg.ValidateBasic(); err != nil {
			return nil, err
		}
	}

	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i)
		nodeDir := filepath.Join(genCfg.OutputDir, nodeDirName)
		cfg.RootDir = nodeDir
		cfg.ChainCfg.SetRoot(filepath.Join(nodeDir, "abci"))

		err := os.MkdirAll(filepath.Join(nodeDir, "abci", "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return nil, err
		}

		err = os.MkdirAll(filepath.Join(nodeDir, "abci", "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return nil, err
		}
		priv := privateKeys[i]
		cfg.AppCfg.PrivateKey = hex.EncodeToString(priv[:])
		// write node key
		pvkeyFile := filepath.Join(cfg.RootDir, PrivKeyFile)
		err = os.WriteFile(pvkeyFile, []byte(cfg.AppCfg.PrivateKey), 0644)
		if err != nil {
			fmt.Println("Error creating private key file: ", err)
			return nil, err
		}

		cfg.ChainCfg.P2P.AddrBookStrict = false
		cfg.ChainCfg.P2P.AllowDuplicateIP = true
	}

	validatorPkeys := privateKeys[:genCfg.NValidators]
	genDoc := config.GenesisDoc(validatorPkeys, chainIDPrefix)

	// write genesis file
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		if err := genDoc.SaveAs(filepath.Join(nodeDir, "abci", "config", "genesis.json")); err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return nil, err
		}
	}

	// Gather persistent peers addresses
	var persistentPeers string

	if genCfg.PopulatePersistentPeers {
		persistentPeers = persistentPeersString(cfg, genCfg, privateKeys)
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
		cfg.AppCfg.PrivateKey = hex.EncodeToString(privateKeys[i])
		cfg.AppCfg.PrivateKeyPath = PrivKeyFile
		writeConfigFile(filepath.Join(nodeDir, "config.toml"), cfg)
	}

	fmt.Printf("Successfully initialized %d node directories: %s\n",
		genCfg.NValidators+genCfg.NNonValidators, genCfg.OutputDir)

	return privateKeys, nil
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
	cfg.AppCfg.PrivateKey = hex.EncodeToString(priv[:])
	nodeID := (&p2p.NodeKey{PrivKey: priv}).ID()
	fmt.Printf("Node Id for node-%d: %v\n", nodeIdx, nodeID)

	// genesis file
	genFile := cfg.ChainCfg.GenesisFile()
	if cmtos.FileExists(genFile) {
		fmt.Printf("Found genesis file %v\n", genFile)
	} else {
		genDoc := config.GenesisDoc([]ed25519.PrivKey{priv}, chainIDPrefix)
		genDoc.Validators[0].Name = fmt.Sprintf("validator-%d", nodeIdx)
		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}

		fmt.Printf("Generated genesis file %v\n", genFile)
	}

	// write node key
	pvkeyFile := filepath.Join(cfg.RootDir, PrivKeyFile)
	err = os.WriteFile(pvkeyFile, []byte(cfg.AppCfg.PrivateKey), 0644)
	if err != nil {
		fmt.Println("Error creating private key file: ", err)
		return err
	}
	cfg.AppCfg.PrivateKeyPath = PrivKeyFile
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

	ip[3] += byte(i)
	return ip.String()
}

func persistentPeersString(cfg *config.KwildConfig, genCfg *TestnetGenerateConfig, PrivKeys []ed25519.PrivKey) string {
	persistentPeers := make([]string, genCfg.NValidators+genCfg.NNonValidators)
	for i := 0; i < genCfg.NValidators+genCfg.NNonValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		cfg.RootDir = nodeDir
		cfg.ChainCfg.SetRoot(filepath.Join(nodeDir, "abci"))
		nodeKey := p2p.NodeKey{PrivKey: PrivKeys[i]}
		nodeID := nodeKey.ID()
		persistentPeers[i] = p2p.IDAddressString(nodeID, fmt.Sprintf("%s:%d", hostnameOrIP(genCfg, i), genCfg.P2pPort))
	}
	return strings.Join(persistentPeers, ",")
}
