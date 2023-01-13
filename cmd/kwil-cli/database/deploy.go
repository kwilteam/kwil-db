package database

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/common/display"
	"kwil/kwil/client/grpc-client"
	execUtils "kwil/x/execution/utils"
	"os"
)

func cmdDeploy(cmd *cobra.Command, args []string) error {
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

		db, err := execUtils.DBFromJson(file)
		if err != nil {
			return fmt.Errorf("failed to parse database: %w", err)
		}

		client, err := grpc_client.NewClient(cc, viper.GetViper())
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		res, err := client.DeployDatabase(cmd.Context(), db)
		if err != nil {
			return err
		}

		display.PrintTxResponse(res)
		return nil
	})
}
