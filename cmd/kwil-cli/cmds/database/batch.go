package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"kwil/pkg/csv"
	"kwil/pkg/engine/types"
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
	var csvColumnMappings []string
	var inputValueMappings []string // these override the csv column mappings
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

				fileType, err := getFileType(filePath)
				if err != nil {
					return fmt.Errorf("error getting file type: %w", err)
				}

				if !isSupportedBatchFileType(fileType) {
					return fmt.Errorf("unsupported file type: %s", fileType)
				}

				file, err := os.Open(filePath)
				if err != nil {
					return fmt.Errorf("error opening file: %w", err)
				}

				inputs, err := buildInputs(file, fileType, csvColumnMappings, inputValueMappings)
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

	cmd.Flags().StringSliceVarP(&csvColumnMappings, "map-input", "m", []string{}, "the variables mappings to the action inputs (e.g. id:$id, name:$name, age:$age)")
	cmd.Flags().StringSliceVarP(&inputValueMappings, "values", "v", []string{}, "the variables mappings to the action inputs (e.g. id:123, name:john, age:25).  These will apply to all rows, and will override the csv column mappings")
	cmd.Flags().StringVarP(&filePath, "path", "p", "", "the path to the file to read in (e.g. /home/user/file.csv)")
	cmd.Flags().StringVarP(&action, "action", "a", "", "the action to execute")
	cmd.Flags().StringP(nameFlag, "n", "", "the database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the database id")

	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("action")
	return cmd
}

// buildInputs builds the inputs for the file
func buildInputs(file *os.File, fileType string, columnMappingFlag []string, inputMappings []string) ([]map[string][]byte, error) {
	switch fileType {
	case "csv":
		return buildCsvInputs(file, columnMappingFlag, inputMappings)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func addInputMappings(inputs []map[string][]byte, inputMappings []string) ([]map[string][]byte, error) {
	for _, inputMapping := range inputMappings {
		parts := strings.SplitN(inputMapping, ":", 2)
		if len(parts) != 2 {
			return inputs, fmt.Errorf("invalid input mapping: %s", inputMapping)
		}

		ensureInputFormat(&parts[0])

		for _, input := range inputs {
			input[parts[0]] = types.NewExplicitMust(parts[1], types.TEXT).Bytes()
		}
	}

	return inputs, nil
}

// buildCsvInputs builds the inputs for a csv file
func buildCsvInputs(file *os.File, columnMappings []string, inputMappings []string) ([]map[string][]byte, error) {
	data, err := csv.Read(file, csv.ContainsHeader)
	if err != nil {
		return nil, fmt.Errorf("error reading csv: %w", err)
	}

	colMappings, err := buildColumnMappings(columnMappings, data.Header)
	if err != nil {
		return nil, fmt.Errorf("error building column mappings: %w", err)
	}

	ins, err := data.BuildInputs(colMappings)
	if err != nil {
		return nil, fmt.Errorf("error building inputs: %w", err)
	}

	ins, err = addInputMappings(ins, inputMappings)
	if err != nil {
		return nil, fmt.Errorf("error adding input mappings: %w", err)
	}

	return ins, nil
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

		ensureInputFormat(&actionInput)

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

		ensureInputFormat(&parts[1])
		res[parts[0]] = parts[1]
	}

	return res, nil
}

func ensureInputFormat(in *string) {
	if !strings.HasPrefix(*in, "$") {
		*in = fmt.Sprintf("$%s", *in)
	}
}

func isSupportedBatchFileType(fileType string) bool {
	for _, supportedType := range supportedBatchFileTypes {
		if supportedType == fileType {
			return true
		}
	}

	return false
}

func getFileType(path string) (string, error) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid file path: %s", path)
	}

	return parts[len(parts)-1], nil
}
