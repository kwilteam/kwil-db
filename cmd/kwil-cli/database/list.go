package database

import (
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/internal/app/kcli"
	"strings"

	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List databases",
		Long: `List lists the databases owned by a wallet.
A wallet can be specified with the --owner flag, otherwise the default wallet is used.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := kcli.New(ctx, common.AppConfig)
			if err != nil {
				return err
			}

			var address string
			// see if they passed an address
			passedAddress, err := cmd.Flags().GetString("owner")
			if err == nil && passedAddress != "NULL" {
				address = passedAddress
			} else {
				// if not, use the default
				address = clt.Config.Fund.GetAccountAddress()
			}

			if address == "" {
				return fmt.Errorf("no address provided")
			}

			dbs, err := clt.Client.ListDatabases(ctx, strings.ToLower(address))
			if err != nil {
				return fmt.Errorf("failed to list databases: %w", err)
			}

			for _, db := range dbs {
				fmt.Println(db)
			}

			return nil
		},
	}

	cmd.Flags().StringP("owner", "o", "NULL", "The owner of the database")
	return cmd
}
