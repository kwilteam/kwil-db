package database

import (
	"encoding/json"
	"fmt"
	"kwil/internal/app/kcli/common/display"
	"kwil/internal/app/kcli/config"
	"kwil/pkg/kwil-client"
	"kwil/x/types/databases"
	"os"

	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy databases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			ctx := cmd.Context()
			clt, err := kwil_client.New(ctx, config.AppConfig)
			if err != nil {
				return err
			}

			res, err := clt.DeployDatabase(ctx, &db)
			if err != nil {
				return err
			}
			fmt.Printf("Deployed database: %+v\n", res)

			display.PrintTxResponse(res)
			return nil
		},
	}

	cmd.Flags().StringP("path", "p", "", "Path to the database definition file (required)")
	cmd.MarkFlagRequired("path")
	return cmd
}
