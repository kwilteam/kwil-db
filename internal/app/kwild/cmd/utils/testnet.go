package utils

import (
	"github.com/kwilteam/kwil-db/internal/pkg/nodecfg"
	"github.com/spf13/cobra"
)

var netFlags nodecfg.TestnetGenerateConfig

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
	return nodecfg.GenerateTestnetConfig(&netFlags)
}

func NewTestnetCmd() *cobra.Command {
	testnetCmd.Flags().IntVarP(&netFlags.NValidators, "validator-num", "v", 4, "number of validators to initialize the testnet with")

	testnetCmd.Flags().IntVarP(&netFlags.NNonValidators, "non-validator-num", "n", 4, "number of non validators to initialize the testnet with")

	testnetCmd.Flags().StringVar(&netFlags.ConfigFile, "config", "", "config file to use (note some options may be overwritten)")

	testnetCmd.Flags().StringVarP(&netFlags.OutputDir, "output-dir", "o", ".testnet", "directory to store initialization data for the testnet")

	testnetCmd.Flags().StringVar(&netFlags.NodeDirPrefix, "node-dir-prefix", "node", "prefix the directory name for each node with (node results in node0, node1, ...)")

	testnetCmd.Flags().Int64Var(&netFlags.InitialHeight, "initial-height", 0, "initial height of the first block")

	testnetCmd.Flags().BoolVar(&netFlags.PopulatePersistentPeers, "populate-persistent-peers", true,
		"update config of each node with the list of persistent peers build using either"+
			" hostname-prefix or starting-ip-address")

	testnetCmd.Flags().IntVar(&netFlags.P2pPort, "p2p-port", 26656, "P2P Port")

	testnetCmd.Flags().StringArrayVar(&netFlags.Hostnames, "hostname", []string{},
		"manually override all hostnames of validators (use --hostname multiple times for multiple hosts)"+
			"Example: --hostname '192.168.10.10' --hostname: '192.168.10.20'")

	testnetCmd.Flags().StringVar(&netFlags.StartingIPAddress, "starting-ip-address", "",
		"starting IP address ("+
			"\"192.168.0.1\""+
			" results in persistent peers list ID0@192.168.0.1:26656, ID1@192.168.0.2:26656, ...)")

	testnetCmd.Flags().StringVar(&netFlags.HostnameSuffix, "hostname-suffix", "",
		"hostname suffix ("+
			"\".xyz.com\""+
			" results in persistent peers list ID0@node0.xyz.com:26656, ID1@node1.xyz.com:26656, ...)")

	testnetCmd.Flags().StringVar(&netFlags.HostnamePrefix, "hostname-prefix", "node",
		"hostname prefix (\"node\" results in persistent peers list ID0@node0:26656, ID1@node1:26656, ...)")

	return testnetCmd
}
