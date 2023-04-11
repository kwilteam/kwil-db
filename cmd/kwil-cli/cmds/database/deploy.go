package database

import (
	"context"
	"encoding/json"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"kwil/pkg/crypto"
	"kwil/pkg/engine/models"
	"kwil/pkg/kl/parser"
	"os"

	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	var filePath string
	var fileType string
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy databases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				// read in the file
				file, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				var db *models.Dataset
				if fileType == "kf" {
					db, err = unmarshalKf(file)
				} else if fileType == "json" {
					db, err = unmarshalJson(file)
				} else {
					return fmt.Errorf("invalid file type: %s", fileType)
				}
				if err != nil {
					return fmt.Errorf("failed to unmarshal file: %w", err)
				}

				db.Owner = crypto.AddressFromPrivateKey(conf.PrivateKey)

				res, err := client.DeployDatabase(ctx, db)
				if err != nil {
					return err
				}

				display.PrintTxResponse(res)
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&filePath, "path", "p", "", "Path to the database definition file (required)")
	cmd.Flags().StringVarP(&fileType, "type", "t", "kf", "File type of the database definition file (kf or json).  defaults to kf (kuneiform).")
	cmd.MarkFlagRequired("path")
	return cmd
}

func unmarshalKf(bts []byte) (*models.Dataset, error) {
	ast, err := parser.Parse(bts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return ast.Dataset(), nil
}

func unmarshalJson(bts []byte) (*models.Dataset, error) {
	var db models.Dataset
	err := json.Unmarshal(bts, &db)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return &db, nil
}
