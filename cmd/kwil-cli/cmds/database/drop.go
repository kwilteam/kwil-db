package database

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"

	"github.com/spf13/cobra"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop DB_NAME",
		Short: "Drops a database",
		Long:  "Drops a database.  Requires 1 argument: the name of the database to drop",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp []byte

			err := common.DialClient(cmd.Context(), 0, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				var _err error
				resp, _err = client.DropDatabase(ctx, args[0])
				if _err != nil {
					return fmt.Errorf("error dropping database: %w", _err)
				}

				return nil
			})

			msg := display.WrapMsg(respTxHash(resp), err)
			return display.Print(msg, err, config.GetOutputFormat())
		},
	}
	return cmd
}
