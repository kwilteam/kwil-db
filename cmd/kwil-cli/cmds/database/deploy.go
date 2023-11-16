package database

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kwilteam/kuneiform/kfparser"
	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
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
			var txHash []byte

			err := common.DialClient(cmd.Context(), 0, func(ctx context.Context, cl *client.Client, conf *config.KwilCliConfig) error {
				// read in the file
				file, err := os.Open(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				defer file.Close()

				var db *transactions.Schema
				if fileType == "kf" {
					db, err = UnmarshalKf(file)
				} else if fileType == "json" {
					db, err = UnmarshalJson(file)
				} else {
					return fmt.Errorf("invalid file type: %s", fileType)
				}
				if err != nil {
					return fmt.Errorf("failed to unmarshal file: %w", err)
				}

				txHash, err = cl.DeployDatabase(ctx, db, client.WithNonce(nonceOverride))
				return err
			})

			return display.Print(display.RespTxHash(txHash), err, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringVarP(&filePath, "path", "p", "", "Path to the database definition file (required)")
	cmd.Flags().StringVarP(&fileType, "type", "t", "kf", "File type of the database definition file (kf or json).  defaults to kf (kuneiform).")
	cmd.MarkFlagRequired("path")
	return cmd
}

func UnmarshalKf(file *os.File) (*transactions.Schema, error) {
	source, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read Kuneiform source file: %w", err)
	}

	astSchema, err := kfparser.Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	schemaJson, err := astSchema.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var db transactions.Schema
	err = json.Unmarshal(schemaJson, &db)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema json: %w", err)
	}

	return &db, nil
}

func UnmarshalJson(file *os.File) (*transactions.Schema, error) {
	bts, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var db transactions.Schema
	err = json.Unmarshal(bts, &db)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return &db, nil
}
