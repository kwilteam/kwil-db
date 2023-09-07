package database

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"

	"github.com/spf13/cobra"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query QUERY_TEXT",
		Short: "Queries a database",
		Long:  "Queries a database. Requires 1 argument: the query text.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp respRelations

			err := common.DialClient(cmd.Context(), common.WithoutPrivateKey,
				func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
					dbid, err := getSelectedDbid(cmd, conf)
					if err != nil {
						return fmt.Errorf("target database not properly specified: %w", err)
					}

					resp.Data, err = client.Query(ctx, dbid, args[0])
					if err != nil {
						return fmt.Errorf("error querying database: %w", err)
					}

					return nil
				})

			msg := display.WrapMsg(&resp, err)
			display.Print(msg, err, config.GetOutputFormat())
			return nil
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the database id")
	return cmd
}
