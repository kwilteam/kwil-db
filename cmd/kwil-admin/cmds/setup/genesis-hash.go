package setup

import (
	"encoding/hex"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/spf13/cobra"
)

var (
	genesisHashLong = `Compute genesis hash from existing PostgreSQL datasets, and optionally update ` + "`" + `genesis.json` + "`" + `.
It takes up to 4 arguments, which are the postgres DB name, host, port, user, and password to access the datasets to be included in the genesis hash.

By default, it will print the genesis hash to stdout. To specify a genesis file to update as well, use the ` + "`" + `--genesis` + "`" + ` flag.`

	genesisHashExample = `# Compute genesis hash from existing PostgreSQL datasets, and add it to a genesis file
kwil-admin setup genesis-hash "kwild" "127.0.0.1" "5432" "kwild" "" --genesis "~/.kwild/abci/config/genesis.json"`
)

func genesisHashCmd() *cobra.Command {

	var genesisFile string

	cmd := &cobra.Command{
		Use:     "genesis-hash",
		Short:   "Compute genesis hash from existing PostgreSQL datasets, and optionally update `genesis.json`.",
		Long:    genesisHashLong,
		Example: genesisHashExample,
		Hidden:  true, // also not listed added, but this is going to be experimental even when implement properly
		Args:    cobra.RangeArgs(0, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			panic("not implemented")
			dbCfg := &pg.ConnConfig{ // defaults
				Host:   "127.0.0.1",
				Port:   "5432",
				DBName: "kwild",
				User:   "kwild",
				Pass:   "kwild",
			}
			if len(args) > 0 {
				dbCfg.Host = args[0]
			}
			if len(args) > 1 {
				dbCfg.Port = args[1]
			}
			if len(args) > 2 {
				dbCfg.DBName = args[2]
			}
			if len(args) > 3 {
				dbCfg.User = args[3]
			}
			if len(args) > 4 {
				dbCfg.Pass = args[4]
			}

			// TODO:
			// appHash, err := config.PatchGenesisAppHash(cmd.Context(), dbCfg, genesisFile)
			// if err != nil {
			// 	return display.PrintErr(cmd, err)
			// }

			var appHash []byte

			return display.PrintCmd(cmd, display.RespString(hex.EncodeToString(appHash)))
		},
	}

	cmd.Flags().StringVarP(&genesisFile, "genesis", "g", "", "optional path to the genesis file to patch with the computed app hash")

	return cmd
}
