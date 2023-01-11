package database

import (
	"context"
	"fmt"
	"kwil/x/cli/chain"
	"kwil/x/cli/client"
	"kwil/x/cli/cmds/display"
	"kwil/x/cli/util"
	"kwil/x/transactions"
	"kwil/x/types/databases"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drops a database",
		Long:  "Drops a database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := client.NewClient(cc, viper.GetViper())
				if err != nil {
					return err
				}
				c, err := chain.NewClientV(viper.GetViper())
				if err != nil {
					return err
				}
				// should be one arg
				if len(args) != 1 {
					return fmt.Errorf("deploy requires one argument: database name")
				}

				data := &databases.DatabaseIdentifier{
					Name:  args[0],
					Owner: c.Address.String(),
				}

				// build tx
				tx, err := client.BuildTransaction(ctx, transactions.DROP_DATABASE, data, c.PrivateKey)
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
