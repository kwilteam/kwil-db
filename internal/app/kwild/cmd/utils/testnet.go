package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	cfg "github.com/cometbft/cometbft/config"
	cmtos "github.com/cometbft/cometbft/libs/os"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	nValidators   int
	initialHeight int64
	configFile    string
	outputDir     string
	nodeDirPrefix string
	disable_gas   bool

	populatePersistentPeers bool
	hostnamePrefix          string
	hostnameSuffix          string
	startingIPAddress       string
	hostnames               []string
	p2pPort                 int
)

const (
	nodeDirPerm = 0755
)

var testnetCmd = &cobra.Command{
	Use:     "testnet",
	Aliases: []string{"net"},
	Short:   "Initializes the files required for a kwil test network",
	Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, env etc.).

Note, strict routability for addresses is turned off in the config file.
Optionally, it will fill in persistent_peers list in config file using either hostnames or IPs.

Example:
	kwild testnet --v 4 --o ./output --populate-persistent-peers --starting-ip-address 192.168.10.2
	`,
	RunE: initTestnet,
}

func initTestnet(cmd *cobra.Command, args []string) error {
	if len(hostnames) > 0 && len(hostnames) != (nValidators) {
		return fmt.Errorf(
			"testnet needs precisely %d hostnames (number of validators) if --hostname parameter is used",
			nValidators,
		)
	}

	config := cfg.DefaultConfig()

	// overwrite default config if set and valid
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
		if err := viper.Unmarshal(config); err != nil {
			return err
		}
		if err := config.ValidateBasic(); err != nil {
			return err
		}
	}

	genVals := make([]types.GenesisValidator, nValidators)

	for i := 0; i < nValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName)
		config.SetRoot(nodeDir)

		err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		err = os.MkdirAll(filepath.Join(nodeDir, "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		if err := InitFilesWithConfig(config); err != nil {
			return err
		}

		pvKeyFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorKey)
		pvStateFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorState)
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

	var chainID string
	if disable_gas {
		chainID = "kwil-chain-gcd-"
	} else {
		chainID = "kwil-chain-gce-"
	}

	genDoc := &types.GenesisDoc{
		ChainID:         chainID + cmtrand.Str(6),
		ConsensusParams: types.DefaultConsensusParams(),
		GenesisTime:     cmttime.Now(),
		InitialHeight:   initialHeight,
		Validators:      genVals,
	}

	// write genesis file
	for i := 0; i < nValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i))
		if err := genDoc.SaveAs(filepath.Join(nodeDir, config.BaseConfig.Genesis)); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
	}

	// Gather persistent peers addresses
	var (
		persistentPeers string
		err             error
	)

	if populatePersistentPeers {
		persistentPeers, err = persistentPeersString(config)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
	}

	// Overwrite default config
	for i := 0; i < nValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i))
		config.SetRoot(nodeDir)
		config.P2P.AddrBookStrict = false
		config.P2P.AllowDuplicateIP = true
		if populatePersistentPeers {
			config.P2P.PersistentPeers = persistentPeers
		}
		cfg.WriteConfigFile(filepath.Join(nodeDir, "config", "config.toml"), config)
	}

	fmt.Printf("Successfully initialized %d node directories\n", nValidators)

	return nil
}

func InitFilesWithConfig(config *cfg.Config) error {
	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	if cmtos.FileExists(privValKeyFile) {
		pv = privval.LoadFilePV(privValKeyFile, privValStateFile)
		fmt.Printf("Found private validator keyfile %v  Statefile: %v\n", privValKeyFile, privValStateFile)
	} else {
		pv = privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		fmt.Printf("Generated private validator keyfile %v  Statefile: %v\n", privValKeyFile, privValStateFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if cmtos.FileExists(nodeKeyFile) {
		fmt.Printf("Found node keyfile %v\n", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		fmt.Printf("Generated node keyfile %v\n", nodeKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
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

	envFile := rootify(".env", config.RootDir)
	if cmtos.FileExists(envFile) {
		fmt.Printf("Found .env file %v\n", envFile)
	} else {
		data := GenerateEnvFileData()
		if err := os.WriteFile(envFile, []byte(data), 0644); err != nil {
			return err
		}
	}

	return nil
}

func hostnameOrIP(i int) string {
	if len(hostnames) > 0 && i < len(hostnames) {
		return hostnames[i]
	}
	if startingIPAddress == "" {
		return fmt.Sprintf("%s%d%s", hostnamePrefix, i, hostnameSuffix)
	}
	ip := net.ParseIP(startingIPAddress)
	ip = ip.To4()
	if ip == nil {
		fmt.Printf("%v: non ipv4 address\n", startingIPAddress)
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

func persistentPeersString(config *cfg.Config) (string, error) {
	persistentPeers := make([]string, nValidators)
	for i := 0; i < nValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i))
		config.SetRoot(nodeDir)
		nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
		if err != nil {
			return "", err
		}
		persistentPeers[i] = p2p.IDAddressString(nodeKey.ID(), fmt.Sprintf("%s:%d", hostnameOrIP(i), p2pPort))
	}
	return strings.Join(persistentPeers, ","), nil
}

func GenerateEnvFileData() string {
	data := "#KWILD_PRIVATE_KEY=\n"
	data += "#KWILD_DEPOSITS_BLOCK_CONFIRMATIONS=1\n"
	data += "#KWILD_DEPOSITS_CHAIN_CODE=2\n"
	data += "#COMET_BFT_HOME = '/app/comet-bft/'\n"
	data += "#KWILD_LOG_LEVEL=debug\n"
	data += "#KWILD_DEPOSITS_CLIENT_CHAIN_RPC_URL=# Example: ws://192.168.1.70:8545\n"
	data += "#KWILD_DEPOSITS_POOL_ADDRESS=# Example: 0xF2Df0b975c0C9eFa2f8CA0491C2d1685104d2488"
	return data
}

func NewTestnetCmd() *cobra.Command {
	testnetCmd.Flags().IntVar(&nValidators, "v", 4, "number of validators to initialize the testnet with")

	testnetCmd.Flags().StringVar(&configFile, "config", "", "config file to use (note some options may be overwritten)")

	testnetCmd.Flags().StringVar(&outputDir, "o", "./mytestnet", "directory to store initialization data for the testnet")

	testnetCmd.Flags().StringVar(&nodeDirPrefix, "node-dir-prefix", "node", "prefix the directory name for each node with (node results in node0, node1, ...)")

	testnetCmd.Flags().Int64Var(&initialHeight, "initial-height", 0, "initial height of the first block")

	testnetCmd.Flags().BoolVar(&disable_gas, "disable-gas", false, "Disables gas costs on all transactions and once the network is initialized, it can't be changed")

	testnetCmd.Flags().BoolVar(&populatePersistentPeers, "populate-persistent-peers", true,
		"update config of each node with the list of persistent peers build using either"+
			" hostname-prefix or starting-ip-address")

	testnetCmd.Flags().IntVar(&p2pPort, "p2p-port", 26656, "P2P Port")

	testnetCmd.Flags().StringArrayVar(&hostnames, "hostname", []string{},
		"manually override all hostnames of validators (use --hostname multiple times for multiple hosts)"+
			"Example: --hostname '192.168.10.10' --hostname: '192.168.10.20'")

	testnetCmd.Flags().StringVar(&startingIPAddress, "starting-ip-address", "",
		"starting IP address ("+
			"\"192.168.0.1\""+
			" results in persistent peers list ID0@192.168.0.1:26656, ID1@192.168.0.2:26656, ...)")

	testnetCmd.Flags().StringVar(&hostnameSuffix, "hostname-suffix", "",
		"hostname suffix ("+
			"\".xyz.com\""+
			" results in persistent peers list ID0@node0.xyz.com:26656, ID1@node1.xyz.com:26656, ...)")

	testnetCmd.Flags().StringVar(&hostnamePrefix, "hostname-prefix", "node",
		"hostname prefix (\"node\" results in persistent peers list ID0@node0:26656, ID1@node1:26656, ...)")

	return testnetCmd
}
