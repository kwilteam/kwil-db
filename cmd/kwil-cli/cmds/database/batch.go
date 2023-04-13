package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"

	"github.com/spf13/cobra"
)

type batchFileType string

const (
	batchFileTypeCSV batchFileType = "csv"
)

// batch is used for batch operations on databases
func batchCmd() *cobra.Command {
	var fileType string

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch executes an action",
		Long: `The batch command is used to batch execute an action on a database.  It
reads in a file from the specified directory, and executes the action in bulk.
The execution is treated as a single transaction, and will either succeed or fail.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				res, err := client.DropDatabase(ctx, args[0])
				if err != nil {
					return fmt.Errorf("error dropping database: %w", err)
				}

				display.PrintTxResponse(res)

				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&fileType, "file-type", "", "csv", "the type of file to read in")
	return cmd
}
