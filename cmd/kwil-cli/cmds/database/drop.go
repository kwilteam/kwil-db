package database

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/spf13/cobra"
)

var (
	dropLong = `Drops a database from the connected network.

The drop coommand will drop a database schema, and all of its data, from the connected network.
This will only work if the wallet address that signs the transaction is the owner of the database.

Drop takes one argument: the name of the database to drop.`

	dropExample = `# Drop a database deployed by the current wallet named "mydb"
kwil-cli database drop mydb`
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "drop <db_name>",
		Short:   "Drops a database from the connected network.",
		Long:    dropLong,
		Example: dropExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl common.Client, conf *config.KwilCliConfig) error {
				var err error
				resp, err := cl.DropDatabase(ctx, args[0], client.WithNonce(nonceOverride),
					client.WithSyncBroadcast(syncBcast))
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error dropping database: %w", err))
				}

				return display.PrintCmd(cmd, display.RespTxHash(resp))
			})
		},
	}
	return cmd
}
