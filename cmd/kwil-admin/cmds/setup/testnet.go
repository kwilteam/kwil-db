package setup

import (
	"time"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	"github.com/spf13/cobra"
)

var (
	testnetLong = `The ` + "`" + `testnet` + "`" + ` command is used to create multiple node configurations, all with the same genesis config,
and pre-configured to connect to each other. It will generate a directory for each node, with the necessary files to run each node.

The config files for each of the nodes will specify all of the other nodes as persistent peers so that they will connect to each other on startup.
This is generally only practical for small test networks with fewer than 12 nodes.

The testnet command creates "v + n" node root directories and populates
each with necessary files to start the new network. The genesis file includes list of v validators under the validators section.

NOTE: strict routability for addresses is turned off in the config file so that
the test network of nodes can run on a LAN.`

	testnetExample = `# Generate a network with 4 validators and 4 non-validators with the IPs
# 192.168.10.{2,...,9}
kwil-admin setup testnet --validators 4 --non-validators 4 --output-dir ~/.kwild-testnet

# Same as above but only 2 additional (non-validator) nodes
kwil-admin setup testnet -v 4 -n 2 --o ./output --starting-ip 192.168.10.2

# Manually specify hostnames for the nodes
kwil-admin setup testnet -v 4 -o ./output --hostnames 192.168.10.2 192.168.10.3 ...`
)

func testnetCmd() *cobra.Command {
	var chainId, configFile, outputDir, hostnamePrefix, hostnameSuffix, startingIPAddress, nodeDirPrefix string
	var hostnames []string
	var blockInterval time.Duration
	var joinExpiry int64
	var validatorAmount, nonValidatorAmount, p2pPort int
	var withGas, snapshotsEnabled, privateMode bool
	var allocs AllocsFlag
	var snapshotHeights, maxSnapshots uint64

	cmd := &cobra.Command{
		Use:     "testnet",
		Short:   "The `testnet` command is used to create multiple node configurations, all with the same genesis config, and pre-configured to connect to each other.",
		Long:    testnetLong,
		Example: testnetExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := nodecfg.GenerateTestnetConfig(&nodecfg.TestnetGenerateConfig{
				ChainID:                 chainId,
				BlockInterval:           blockInterval,
				NValidators:             validatorAmount,
				NNonValidators:          nonValidatorAmount,
				ConfigFile:              configFile,
				OutputDir:               outputDir,
				NodeDirPrefix:           nodeDirPrefix,
				PopulatePersistentPeers: true,
				HostnamePrefix:          hostnamePrefix,
				HostnameSuffix:          hostnameSuffix,
				StartingIPAddress:       startingIPAddress,
				Hostnames:               hostnames,
				P2pPort:                 p2pPort,
				JoinExpiry:              joinExpiry,
				WithoutGasCosts:         !withGas,
				Allocs:                  allocs.M,
				SnapshotsEnabled:        snapshotsEnabled,
				MaxSnapshots:            maxSnapshots,
				SnapshotHeights:         snapshotHeights,
				PrivateMode:             privateMode,
			}, &nodecfg.ConfigOpts{
				UniquePorts: true,
			})
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "./.testnet", "parent directory for all of generated node folders")
	cmd.Flags().StringVar(&configFile, "config", "", "path to a config file to use as a template for all nodes")
	cmd.Flags().StringVar(&chainId, "chain-id", "", "chain ID to use for the genesis file")
	cmd.Flags().StringVar(&hostnamePrefix, "hostname-prefix", "", "prefix for hostnames of nodes")
	cmd.Flags().StringVar(&hostnameSuffix, "hostname-suffix", "", "suffix for hostnames of nodes")
	cmd.Flags().StringVar(&nodeDirPrefix, "node-dir-prefix", "node", "prefix for the node directories (node results in node0, node1, ...)")
	cmd.Flags().StringVar(&startingIPAddress, "starting-ip", "172.10.100.2", "starting IP address for nodes")
	cmd.Flags().StringSliceVar(&hostnames, "hostnames", []string{}, "override all hostnames of the nodes (list of hostnames must be the same length as the number of nodes)")
	cmd.Flags().IntVarP(&p2pPort, "p2p-port", "p", 26656, "p2p port for nodes")
	cmd.Flags().DurationVarP(&blockInterval, "block-interval", "i", 6*time.Second, "shortest block interval in seconds")
	cmd.Flags().Int64Var(&joinExpiry, "join-expiry", 100800, "number of blocks before a join request expires")
	cmd.Flags().IntVarP(&validatorAmount, "validators", "v", 3, "number of validators to generate")
	cmd.Flags().IntVarP(&nonValidatorAmount, "non-validators", "n", 0, "number of non-validators to generate")
	cmd.Flags().BoolVar(&withGas, "gas", false, "enable gas")
	cmd.Flags().Var(&allocs, "alloc", "account=amount pairs of genesis account allocations")
	cmd.Flags().BoolVar(&snapshotsEnabled, "snaps", false, "enables db snapshots")
	cmd.Flags().Uint64Var(&maxSnapshots, "max-snaps", 3, "maximum number of snapshots to store in the device")
	cmd.Flags().Uint64Var(&snapshotHeights, "snap-heights", 10000, "recurring heights (multipes of --snap-heights) to take snapshots at")
	cmd.Flags().BoolVarP(&privateMode, "private", "m", false, "enable private mode for the network")
	return cmd
}
