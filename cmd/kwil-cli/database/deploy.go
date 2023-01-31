package database

import (
	"context"
	"encoding/json"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/common/display"
	"kwil/pkg/grpc/client"
	"kwil/x/fund"
	"kwil/x/types/databases"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
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

				conf, err := fund.NewConfig()
				if err != nil {
					return fmt.Errorf("error getting client config: %w", err)
				}

				client, err := client.NewClient(cc, conf)
				if err != nil {
					return fmt.Errorf("failed to create client: %w", err)
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
