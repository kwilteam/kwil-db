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

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List databases",
		Long: `"List" lists the databases owned by a wallet.
A wallet can be specified with the --owner flag, otherwise the default wallet is used.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp *respDBList
			err := common.DialClient(cmd.Context(), common.WithoutPrivateKey, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				owner, err := getSelectedOwner(cmd, conf)
				if err != nil {
					return err
				}

				dbs, err := client.ListDatabases(ctx, owner)
				if err != nil {
					return fmt.Errorf("failed to list databases: %w", err)
				}

				resp = &respDBList{
					Databases: dbs,
					Owner:     owner,
				}

				return nil
			})

			msg := display.WrapMsg(resp, err)
			return display.Print(msg, err, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringP(ownerFlag, "o", "", "The owner of the database")
	return cmd
}
