package database

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/spf13/cobra"
)

var (
	deployLong = `Deploy a database schema to the target Kwil node.
A path to a file containing the database schema must be provided as the first positional argument.

Either a Kuneiform or a JSON file can be provided.  The file type is determined by the --type flag.
By default, the file type is kf (Kuneiform).  Pass --type json to deploy a JSON file.`

	deployExample = `# Deploy a database schema to the target Kwil node
kwil-cli database deploy ./schema.kf`
)

func deployCmd() *cobra.Command {
	var filePath, fileType, overrideName string

	cmd := &cobra.Command{
		Use:     "deploy <path>",
		Short:   "Deploy a database schema to the target Kwil node.",
		Long:    deployLong,
		Example: deployExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				if cmd.Flags().Changed("path") {
					if len(args) > 0 {
						return display.PrintErr(cmd, fmt.Errorf("no positional arguments are allowed when using the --path flag"))
					}
				} else {
					if len(args) == 0 {
						return display.PrintErr(cmd, fmt.Errorf("must provide a path to the database schema file"))
					}
					filePath = args[0]
				}

				// read in the file
				file, err := os.Open(filePath)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("failed to read file: %w", err))
				}
				defer file.Close()

				var db *types.Schema
				if fileType == "kf" {
					db, err = UnmarshalKf(file)
				} else if fileType == "json" {
					db, err = UnmarshalJson(file)
				} else {
					return display.PrintErr(cmd, fmt.Errorf("invalid file type: %s", fileType))
				}
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("failed to unmarshal file: %w", err))
				}

				if cmd.Flags().Changed("name") {
					if overrideName == "" {
						return display.PrintErr(cmd, fmt.Errorf("--name flag cannot be empty string"))
					}
					db.Name = overrideName
				}

				txHash, err := cl.DeployDatabase(ctx, db, clientType.WithNonce(nonceOverride),
					clientType.WithSyncBroadcast(syncBcast))
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("failed to deploy database: %w", err))
				}
				// If sycnBcast, and we have a txHash (error or not), do a query-tx.
				if len(txHash) != 0 && syncBcast {
					time.Sleep(500 * time.Millisecond) // otherwise it says not found at first
					resp, err := cl.TxQuery(ctx, txHash)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("tx query failed: %w", err))
					}
					return display.PrintCmd(cmd, display.NewTxHashAndExecResponse(resp))
				}
				return display.PrintCmd(cmd, display.RespTxHash(txHash))
			})
		},
	}

	cmd.Flags().StringVarP(&filePath, "path", "p", "", "path to the database definition file (required)")
	err := cmd.Flags().MarkDeprecated("path", "specify the path as the first positional argument instead")
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&fileType, "type", "t", "kf", "file type of the database definition file (kf or json)")
	cmd.Flags().StringVarP(&overrideName, "name", "n", "", "set the name of the database, overriding the name in the schema file")

	cmd.MarkFlagRequired("path")
	return cmd
}

func UnmarshalKf(file *os.File) (*types.Schema, error) {
	source, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read Kuneiform source file: %w", err)
	}

	return parse.Parse(source)
}

func UnmarshalJson(file *os.File) (*types.Schema, error) {
	bts, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var db types.Schema
	err = json.Unmarshal(bts, &db)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return &db, nil
}
