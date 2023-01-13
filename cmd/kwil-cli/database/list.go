package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	grpc_client "kwil/kwil/client/grpc-client"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List databases",
		Long: `List lists the databases owned by a wallet.
A wallet can be specified with the --owner flag, otherwise the default wallet is used.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := grpc_client.NewClient(cc, viper.GetViper())
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
					address = client.Config.Address
				}

				if address == "" {
					return fmt.Errorf("no address provided")
				}

				dbs, err := client.Txs.ListDatabases(ctx, strings.ToLower(address))
				if err != nil {
					return fmt.Errorf("failed to list databases: %w", err)
				}

				for _, db := range dbs {
					fmt.Println(db)
				}

				return nil
			})
		},
	}

	cmd.Flags().StringP("owner", "o", "NULL", "The owner of the database")
	return cmd
}
