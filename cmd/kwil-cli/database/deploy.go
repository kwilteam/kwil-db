package database

import (
	"context"
	"encoding/json"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/common/display"
	grpc_client "kwil/kwil/client/grpc-client"
	"kwil/x/types/databases"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				if len(args) != 0 {
					return fmt.Errorf("deploy command does not take any arguments")
				}

				filePath, err := cmd.Flags().GetString("path")
				if err != nil {
					return fmt.Errorf("must specify a path path with the --path flag")
				}

				// read in the file
				file, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				var db databases.Database[[]byte]
				err = json.Unmarshal(file, &db)
				if err != nil {
					return fmt.Errorf("failed to unmarshal file: %w", err)
				}

				client, err := grpc_client.NewClient(cc, viper.GetViper())
				if err != nil {
					return fmt.Errorf("failed to create grpc client: %w", err)
				}

				res, err := client.DeployDatabase(cmd.Context(), &db)
				if err != nil {
					return err
				}

				display.PrintTxResponse(res)
				return nil
			})
		},
	}

	cmd.Flags().StringP("path", "p", "", "Path to the database definition file")
	cmd.MarkFlagRequired("path")
	return cmd
}
