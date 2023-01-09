package database

import (
	"context"
	"fmt"
	"kwil/x/cli/chain"
	"kwil/x/cli/client"
	"kwil/x/cli/cmds/display"
	"kwil/x/cli/util"
	"kwil/x/execution/clean"
	execUtils "kwil/x/execution/utils"
	"kwil/x/execution/validation"
	"kwil/x/transactions"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a database",
		Long:  "Deploy a database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := client.NewClient(cc, viper.GetViper())
				if err != nil {
					return err
				}
				// should be one arg
				if len(args) != 1 {
					return fmt.Errorf("deploy requires one argument: path")
				}
				c, err := chain.NewClientV(viper.GetViper())
				if err != nil {
					return err
				}

				cx := viper.GetString("chain-code")
				fmt.Printf("Chain code: %s", cx)

				// read in the file
				file, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				db, err := execUtils.DBFromJson(file)
				if err != nil {
					return fmt.Errorf("failed to parse database: %w", err)
				}

				clean.CleanDatabase(db)

				// validate the database
				err = validation.ValidateDatabase(db)
				if err != nil {
					return fmt.Errorf("error on database: %w", err)
				}

				if db.Owner != c.Address.String() {
					return fmt.Errorf("database owner must be the same as the current account")
				}

				// build tx
				tx, err := client.BuildTransaction(ctx, transactions.DEPLOY_DATABASE, db, c.PrivateKey)
				if err != nil {
					return err
				}

				res, err := client.Broadcast(ctx, tx)
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
