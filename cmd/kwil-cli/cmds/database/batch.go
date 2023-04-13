package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"kwil/pkg/csv"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	supportedBatchFileTypes = []string{"csv"}
)

// batch is used for batch operations on databases
func batchCmd() *cobra.Command {
	var filePath string
	var fileType string
	var csvColumnMappings []string
	var action string

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch executes an action",
		Long: `The batch command is used to batch execute an action on a database.  It
reads in a file from the specified directory, and executes the action in bulk.
The execution is treated as a single transaction, and will either succeed or fail.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				dbid, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return err
				}

				if !isSupportedBatchFileType(fileType) {
					return fmt.Errorf("unsupported file type: %s", fileType)
				}

				file, err := os.Open(filePath)
				if err != nil {
					return fmt.Errorf("error opening file: %w", err)
				}

				inputs, err := buildInputs(file, fileType, csvColumnMappings)
				if err != nil {
					return fmt.Errorf("error building inputs: %w", err)
				}

				receipt, _, err := client.ExecuteActionSerialized(ctx, dbid, strings.ToLower(action), inputs)
				if err != nil {
					return fmt.Errorf("error executing action: %w", err)
				}

				// print the response
				display.PrintTxResponse(receipt)

				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&fileType, "file-type", "t", "csv", "the type of file to read in")
	cmd.Flags().StringSliceVarP(&csvColumnMappings, "map-input", "m", []string{}, "the variables mappings to the action inputs (e.g. id:$id, name:$name, age:$age)")
	cmd.Flags().StringVarP(&filePath, "path", "p", "", "the path to the file to read in (e.g. /home/user/file.csv)")
	cmd.Flags().StringVarP(&action, "action", "a", "", "the action to execute")
	cmd.Flags().StringP(nameFlag, "n", "", "the database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the database id")

	cmd.MarkFlagRequired("file-type")
	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("action")
	return cmd
}

// buildInputs builds the inputs for the file
func buildInputs(file *os.File, fileType string, columnMappingFlag []string) ([]map[string][]byte, error) {
	switch fileType {
	case "csv":
		return buildCsvInputs(file, columnMappingFlag)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}

// buildCsvInputs builds the inputs for a csv file
func buildCsvInputs(file *os.File, columnMappings []string) ([]map[string][]byte, error) {
	data, err := csv.Read(file, csv.ContainsHeader)
	if err != nil {
		return nil, fmt.Errorf("error reading csv: %w", err)
	}

	colMappings, err := buildColumnMappings(columnMappings, data.Header)
	if err != nil {
		return nil, fmt.Errorf("error building column mappings: %w", err)
	}

	return data.BuildInputs(colMappings)
}

// buildColumnMappings builds the map used to map columns to inputs
// if the mapping provided is empty, it will use the column name as the input name
// if will dynamically add the $ to the input name if it is not provided
func buildColumnMappings(mappings []string, headers []string) (map[string]string, error) {
	if len(mappings) > 0 {
		return convertColumnMappings(mappings)
	}

	return convertHeadersToColumnMappings(headers), nil
}

func convertHeadersToColumnMappings(headers []string) map[string]string {
	res := make(map[string]string)

	for _, header := range headers {
		actionInput := header
		if !strings.HasPrefix(header, "$") {
			actionInput = fmt.Sprintf("$%s", header)
		}

		res[header] = actionInput
	}

	return res
}

// convertColumnMappings converts a list of mappings in the form of "id:$id" to a map
func convertColumnMappings(mappings []string) (map[string]string, error) {
	res := make(map[string]string)

	for _, mapping := range mappings {
		parts := strings.Split(mapping, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid mapping: %s", mapping)
		}

		res[parts[0]] = parts[1]
	}

	return res, nil
}

func isSupportedBatchFileType(fileType string) bool {
	for _, supportedType := range supportedBatchFileTypes {
		if supportedType == fileType {
			return true
		}
	}

	return false
}
