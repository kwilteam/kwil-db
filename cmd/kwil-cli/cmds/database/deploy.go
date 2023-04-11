package database

import (
	"context"
	"encoding/json"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"kwil/pkg/engine/models"
	"os"

	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy databases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				filePath, err := cmd.Flags().GetString("path")
				if err != nil {
					return fmt.Errorf("must specify a path path with the --path flag")
				}

				// read in the file
				file, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				var db models.Dataset
				err = json.Unmarshal(file, &db)
				if err != nil {
					return fmt.Errorf("failed to unmarshal file: %w", err)
				}

				res, err := client.DeployDatabase(ctx, &db)
				if err != nil {
					return err
				}

				display.PrintTxResponse(res)
				return nil
			})
		},
	}

	cmd.Flags().StringP("path", "p", "", "Path to the database definition file (required)")
	cmd.MarkFlagRequired("path")
	return cmd
}
