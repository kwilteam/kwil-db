package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/common/display"
	grpc_client "kwil/kwil/client/grpc-client"
	"kwil/x/types/databases"
	"kwil/x/types/transactions"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drops a database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := grpc_client.NewClient(cc, viper.GetViper())
				if err != nil {
					return err
				}
				// should be one arg
				if len(args) != 1 {
					return fmt.Errorf("deploy requires one argument: database name")
				}

				data := &databases.DatabaseIdentifier{
					Name:  args[0],
					Owner: client.Config.Address,
				}

				// build tx
				tx, err := client.BuildTransaction(ctx, transactions.DROP_DATABASE, data, client.Config.PrivateKey)
				if err != nil {
					return err
				}

				res, err := client.Txs.Broadcast(ctx, tx)
				if err != nil {
					return err
				}

				display.PrintTxResponse(res)

				return nil
			})
		},
	}
	return cmd
}
