package nodecfg

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	cmtCfg "github.com/cometbft/cometbft/config"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/p2p"

	cmtrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"

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

// TODO: if we use our own keys for cosmos, this will not work
// privval.LoadFilePV will need to be replacew with something else
func GenerateNodeConfig(genCfg *NodeGenerateConfig) error {
	cfg := cmtCfg.DefaultConfig()
	cfg.SetRoot(genCfg.HomeDir)
	err := os.MkdirAll(filepath.Join(genCfg.HomeDir, "config"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(genCfg.HomeDir)
		return err
	}

	err = os.MkdirAll(filepath.Join(genCfg.HomeDir, "data"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(genCfg.HomeDir)
		return err
	}
	err = initFilesWithConfig(cfg)
	if err != nil {
		return err
	}
	pvKeyFile := filepath.Join(genCfg.HomeDir, cfg.BaseConfig.PrivValidatorKey)
	pvStateFile := filepath.Join(genCfg.HomeDir, cfg.BaseConfig.PrivValidatorState)
	pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

	pubKey, err := pv.GetPubKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	genVal := types.GenesisValidator{
		Address: pubKey.Address(),
		PubKey:  pubKey,
		Power:   1,
		Name:    "node-0",
	}

	vals := []types.GenesisValidator{genVal}

	genDoc := &types.GenesisDoc{
		ChainID:         chainIDPrefix + cmtrand.Str(6),
		ConsensusParams: types.DefaultConsensusParams(),
		GenesisTime:     cmttime.Now(),
		InitialHeight:   genCfg.InitialHeight,
		Validators:      vals,
	}

	if err := genDoc.SaveAs(filepath.Join(genCfg.HomeDir, cfg.BaseConfig.Genesis)); err != nil {
		_ = os.RemoveAll(genCfg.HomeDir)
		return err
	}

	cfg.P2P.AddrBookStrict = false
	cfg.P2P.AllowDuplicateIP = true
	cfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	cmtCfg.WriteConfigFile(filepath.Join(genCfg.HomeDir, "config", "config.toml"), cfg)
	return nil
}

func GenerateTestnetConfig(genCfg *TestnetGenerateConfig) error {
	if len(genCfg.Hostnames) > 0 && len(genCfg.Hostnames) != (genCfg.NValidators) {
		return fmt.Errorf(
			"testnet needs precisely %d hostnames (number of validators) if --hostname parameter is used",
			genCfg.NValidators,
		)
	}

	cfg := cmtCfg.DefaultConfig()

	// overwrite default config if set and valid
	if genCfg.ConfigFile != "" {
		viper.SetConfigFile(genCfg.ConfigFile)
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
		if err := viper.Unmarshal(cfg); err != nil {
			return err
		}
		if err := cfg.ValidateBasic(); err != nil {
			return err
		}
	}

	genVals := make([]types.GenesisValidator, genCfg.NValidators)

	for i := 0; i < genCfg.NValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i)
		nodeDir := filepath.Join(genCfg.OutputDir, nodeDirName)
		cfg.SetRoot(nodeDir)

		err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}

		err = os.MkdirAll(filepath.Join(nodeDir, "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}

		if err := initFilesWithConfig(cfg); err != nil {
			return err
		}

		pvKeyFile := filepath.Join(nodeDir, cfg.BaseConfig.PrivValidatorKey)
		pvStateFile := filepath.Join(nodeDir, cfg.BaseConfig.PrivValidatorState)
		pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("failed to get public key: %w", err)
		}

		genVals[i] = types.GenesisValidator{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   1,
			Name:    fmt.Sprintf("validator-%d", i),
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
	for i := 0; i < genCfg.NValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		if err := genDoc.SaveAs(filepath.Join(nodeDir, cfg.BaseConfig.Genesis)); err != nil {
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
		persistentPeers, err = persistentPeersString(cfg, genCfg)
		if err != nil {
			_ = os.RemoveAll(genCfg.OutputDir)
			return err
		}
	}

	// Overwrite default config
	for i := 0; i < genCfg.NValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		cfg.SetRoot(nodeDir)
		cfg.P2P.AddrBookStrict = false
		cfg.P2P.AllowDuplicateIP = true
		if genCfg.PopulatePersistentPeers {
			cfg.P2P.PersistentPeers = persistentPeers
		}
		cmtCfg.WriteConfigFile(filepath.Join(nodeDir, "config", "config.toml"), cfg)
	}

	fmt.Printf("Successfully initialized %d node directories\n", genCfg.NValidators)

	return nil
}

// TODO: we definitely want to get rid of this, or at least make it more understandable / move it
// It generates private keys, which we should not leave up to Comet
func initFilesWithConfig(cfg *cmtCfg.Config) error {
	// private validator
	privValKeyFile := cfg.PrivValidatorKeyFile()
	privValStateFile := cfg.PrivValidatorStateFile()
	var pv *privval.FilePV
	if cmtos.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
		fmt.Printf("Found private validator keyfile %v \n		Statefile: %v\n",
			privValKeyFile, privValStateFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		fmt.Printf("Generated private validator keyfile %v \n		Statefile: %v\n",
			privValKeyFile, privValStateFile)
	}

	nodeKeyFile := cfg.NodeKeyFile()
	if cmtos.FileExists(nodeKeyFile) {
		fmt.Printf("Found node keyfile %v\n", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		fmt.Printf("Generated node keyfile %v\n", nodeKeyFile)
	}

	// genesis file
	genFile := cfg.GenesisFile()
	if cmtos.FileExists(genFile) {
		fmt.Printf("Found genesis file %v\n", genFile)
	} else {
		genDoc := types.GenesisDoc{
			ChainID:         fmt.Sprintf("test-chain-%v", cmtrand.Str(6)),
			GenesisTime:     cmttime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		}
		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("can't get pubkey: %w", err)
		}
		genDoc.Validators = []types.GenesisValidator{{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		fmt.Printf("Generated genesis file %v\n", genFile)
	}

	envFile := rootify(".env", cfg.RootDir)
	if cmtos.FileExists(envFile) {
		fmt.Printf("Found .env file %v\n", envFile)
	} else {
		data := generateEnv()
		if err := os.WriteFile(envFile, []byte(data), 0644); err != nil {
			return err
		}
	}

	return nil
}

func generateEnv() string {
	data := "#KWILD_PRIVATE_KEY=\n"
	data += "#KWILD_DEPOSITS_BLOCK_CONFIRMATIONS=1\n"
	data += "#KWILD_DEPOSITS_CHAIN_CODE=2\n"
	data += "#COMET_BFT_HOME = '/app/comet-bft/'\n"
	data += "#KWILD_LOG_LEVEL=debug\n"
	data += "#KWILD_DEPOSITS_CLIENT_CHAIN_RPC_URL=# Example: ws://192.168.1.70:8545\n"
	data += "#KWILD_DEPOSITS_POOL_ADDRESS=# Example: 0xF2Df0b975c0C9eFa2f8CA0491C2d1685104d2488"
	return data
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

func rootify(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func persistentPeersString(cfg *cmtCfg.Config, genCfg *TestnetGenerateConfig) (string, error) {
	persistentPeers := make([]string, genCfg.NValidators)
	for i := 0; i < genCfg.NValidators; i++ {
		nodeDir := filepath.Join(genCfg.OutputDir, fmt.Sprintf("%s%d", genCfg.NodeDirPrefix, i))
		cfg.SetRoot(nodeDir)
		nodeKey, err := p2p.LoadNodeKey(cfg.NodeKeyFile())
		if err != nil {
			return "", err
		}
		persistentPeers[i] = p2p.IDAddressString(nodeKey.ID(), fmt.Sprintf("%s:%d", hostnameOrIP(genCfg, i), genCfg.P2pPort))
	}
	return strings.Join(persistentPeers, ","), nil
}
