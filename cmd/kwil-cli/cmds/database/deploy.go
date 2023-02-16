package database

import (
	"encoding/json"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"kwil/pkg/databases"
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
			clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl,
				client.WithoutServiceConfig(),
			)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			ecdsaKey, err := config.GetEcdsaPrivateKey()
			if err != nil {
				return fmt.Errorf("failed to get ecdsa key: %w", err)
			}

			res, err := clt.DeployDatabase(ctx, &db, ecdsaKey)
			if err != nil {
				return err
			}

			display.PrintTxResponse(res)
			return nil
		},
	}

	cmd.Flags().StringP("path", "p", "", "Path to the database definition file (required)")
	cmd.MarkFlagRequired("path")
	return cmd
}
