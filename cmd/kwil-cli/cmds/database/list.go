package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"

	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List databases",
		Long: `"List" lists the databases owned by a wallet.
A wallet can be specified with the --owner flag, otherwise the default wallet is used.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				owner, err := getSelectedOwner(cmd, conf)
				if err != nil {
					return err
				}

				dbs, err := client.ListDatabases(ctx, strings.ToLower(owner))
				if err != nil {
					return fmt.Errorf("failed to list databases: %w", err)
				}

				if len(dbs) == 0 {
					fmt.Printf("No databases found for address '%s'.\n", owner)
				} else {
					fmt.Printf("Databases belonging to '%s':\n", owner)
				}
				for _, db := range dbs {
					fmt.Println(" - " + db + "   (dbid:" + utils.GenerateDBID(db, owner) + ")")
				}

				return nil
			})
		},
	}

	cmd.Flags().StringP(ownerFlag, "o", "", "The owner of the database")
	return cmd
}
