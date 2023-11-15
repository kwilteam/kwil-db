package setup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/spf13/cobra"
)

var (
	peerLong = `The ` + "`" + `peer` + "`" + ` command facilitates quick setup of a Kwil node as a peer to an existing node.
It will automatically generate required directories and keypairs, and can be given a genesis file and peer list for an existing network.`

	peerExample = `# Initialize a node as a peer to an existing network
` // TODO: add example
)

func peerCmd() *cobra.Command {
	var out, genesisPath string
	var peers []string

	cmd := &cobra.Command{
		Use:     "peer",
		Short:   "The `peer` command facilitates quick setup of a Kwil node as a peer to an existing node.",
		Long:    peerLong,
		Example: peerExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			expandedDir, err := expandPath(out)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			err = os.MkdirAll(expandedDir, 0755)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// if genesis path is given, copy it to the node directory
			if genesisPath != "" {
				genesisPath, err = expandPath(genesisPath)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				file, err := os.ReadFile(genesisPath)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				err = os.WriteFile(filepath.Join(expandedDir, cometbft.GenesisJSONName), file, 0644)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
			}

			cleanedPeers := make([]string, 0)
			for _, peer := range peers {
				cleanedPeers = append(cleanedPeers, strings.TrimSpace(peer))
			}

			cfg := config.EmptyConfig()
			cfg.ChainCfg.P2P.PersistentPeers = strings.Join(cleanedPeers, ",")

			_, err = nodecfg.GenerateNodeFiles(expandedDir, cfg)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&out, "output-dir", "o", "./kwild-node", "generated node parent directory [default: ./kwild-node]")
	cmd.Flags().StringVarP(&genesisPath, "genesis", "g", "", "path to genesis file")
	cmd.Flags().StringSliceVarP(&peers, "peer", "p", nil, "peer to connect to (may be given multiple times, or as a comma-separated list)")

	return cmd
}
