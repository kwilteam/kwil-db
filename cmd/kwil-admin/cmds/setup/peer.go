package setup

import (
	"os"
	"path/filepath"

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
kwil-admin setup peer --root-dir ./kwil-node --genesis /path/to/genesis.json`
)

func peerCmd() *cobra.Command {
	cfg := config.DefaultConfig()
	var genesisPath string

	cmd := &cobra.Command{
		Use:     "peer",
		Short:   "The `peer` command facilitates quick setup of a Kwil node as a peer to an existing node.",
		Long:    peerLong,
		Example: peerExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			expandedDir, err := expandPath(cfg.RootDir)
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

			_, err = nodecfg.GenerateNodeFiles(expandedDir, cfg, true)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&genesisPath, "genesis", "g", "", "path to genesis file")
	config.AddConfigFlags(cmd.Flags(), cfg)

	return cmd
}
