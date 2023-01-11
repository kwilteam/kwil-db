package database

import (
	"context"
	"fmt"
	"kwil/x/cli/client"
	"kwil/x/cli/util"
	"kwil/x/cli/util/display"
	"kwil/x/execution/clean"
	execUtils "kwil/x/execution/utils"
	"kwil/x/execution/validation"
	"kwil/x/transactions"
	"os"
	"strings"

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
				if len(args) != 0 {
					return fmt.Errorf("deploy command does not take any arguments")
				}

				filePath, err := cmd.Flags().GetString("path")
				if err != nil {
					return fmt.Errorf("must specify a path path with the --path flag")
				}

				client, err := client.NewClient(cc, viper.GetViper())
				if err != nil {
					return fmt.Errorf("failed to create client: %w", err)
				}

				// read in the file
				file, err := os.ReadFile(filePath)
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

				if !strings.EqualFold(db.Owner, client.Config.Address) {
					return fmt.Errorf("database owner must be the same as the current account.  Owner: %s, Account: %s", db.Owner, client.Config.Address)
				}

				// build tx
				tx, err := client.BuildTransaction(ctx, transactions.DEPLOY_DATABASE, db, client.Config.PrivateKey)
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
