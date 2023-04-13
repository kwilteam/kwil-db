package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"

	"github.com/spf13/cobra"
)

// batch is used for batch operations on databases
func batchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Drops a database",
		Long:  "Drops a database.  Requires 1 argument: the name of the database to drop",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				res, err := client.DropDatabase(ctx, args[0])
				if err != nil {
					return fmt.Errorf("error dropping database: %w", err)
				}

				display.PrintTxResponse(res)

				return nil
			})
		},
	}
	return cmd
}
