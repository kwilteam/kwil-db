package database

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
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
			return helpers.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				var err error
				txHash, err := cl.DropDatabase(ctx, args[0], clientType.WithNonce(nonceOverride),
					clientType.WithSyncBroadcast(syncBcast))
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error dropping database: %w", err))
				}
				// If sycnBcast, and we have a txHash (error or not), do a query-tx.
				if len(txHash) != 0 && syncBcast {
					time.Sleep(500 * time.Millisecond) // otherwise it says not found at first
					resp, err := cl.TxQuery(ctx, txHash)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("tx query failed: %w", err))
					}
					return display.PrintCmd(cmd, display.NewTxHashAndExecResponse(resp))
				}
				return display.PrintCmd(cmd, display.RespTxHash(txHash))
			})
		},
	}
	return cmd
}
