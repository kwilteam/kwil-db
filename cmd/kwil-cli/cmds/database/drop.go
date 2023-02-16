package database

import (
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"

	"github.com/spf13/cobra"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop db_name",
		Short: "Drops a database",
		Long:  "Drops a database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl,
				client.WithoutServiceConfig(),
			)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			ecdsaPk, err := config.GetEcdsaPrivateKey()
			if err != nil {
				return fmt.Errorf("failed to get ecdsa key: %w", err)
			}

			res, err := clt.DropDatabase(ctx, args[0], ecdsaPk)
			if err != nil {
				return fmt.Errorf("error dropping database: %w", err)
			}

			display.PrintTxResponse(res)

			return nil
		},
	}
	return cmd
}
