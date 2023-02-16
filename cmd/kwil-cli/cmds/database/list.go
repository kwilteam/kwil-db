package database

import (
	"fmt"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"strings"

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
			ctx := cmd.Context()

			clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl,
				client.WithoutServiceConfig(),
			)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			owner, err := getSelectedOwner(cmd)
			if err != nil {
				return err
			}

			dbs, err := clt.ListDatabases(ctx, strings.ToLower(owner))
			if err != nil {
				return fmt.Errorf("failed to list databases: %w", err)
			}

			if len(dbs) == 0 {
				fmt.Printf("No databases found for address '%s'.\n", owner)
			} else {
				fmt.Printf("Databases belonging to '%s':\n", owner)
			}
			for _, db := range dbs {
				fmt.Println(" - " + db)
			}

			return nil
		},
	}

	cmd.Flags().StringP(ownerFlag, "o", "", "The owner of the database")
	return cmd
}
