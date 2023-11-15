package setup

import (
	"encoding/hex"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/spf13/cobra"
)

var (
	genesisHashLong = `Compute genesis hash from SQLite files, and optionally update ` + "`" + `genesis.json` + "`" + `.
It takes one argument, which is the path containing the SQLite files to be included in the genesis hash.

By default, it will print the genesis hash to stdout. To specify a genesis file to update as well, use the ` + "`" + `--genesis` + "`" + ` flag.`

	genesisHashExample = `# Compute genesis hash from SQLite files, and add it to a genesis file
kwil-admin setup genesis-hash "~/.kwild/data/kwil.db" --genesis "~/.kwild/abci/config/genesis.json"`
)

func genesisHashCmd() *cobra.Command {

	var genesisFile string

	cmd := &cobra.Command{
		Use:     "genesis-hash",
		Short:   "Compute genesis hash from SQLite files, and optionally update `genesis.json`.",
		Long:    genesisHashLong,
		Example: genesisHashExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			expandedPath, err := expandPath(args[0])
			if err != nil {
				return err
			}

			appHash, err := config.PatchGenesisAppHash(expandedPath, genesisFile)
			if err != nil {
				return err
			}

			return display.PrintCmd(cmd, display.RespString(hex.EncodeToString(appHash)))
		},
	}

	cmd.Flags().StringVarP(&genesisFile, "genesis", "g", "", "optional path to the genesis file to patch with the computed app hash")

	return cmd
}
