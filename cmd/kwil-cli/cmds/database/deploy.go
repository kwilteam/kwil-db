package database

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kwilteam/kuneiform/kfparser"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/serialize"
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
			return common.DialClient(cmd.Context(), 0, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				// read in the file
				file, err := os.Open(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				defer file.Close()

				var db *serialize.Schema
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

func UnmarshalKf(file *os.File) (*serialize.Schema, error) {
	source, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read Kuneiform source file: %w", err)
	}

	astSchema, err := kfparser.Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	schemaJson, err := json.Marshal(astSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var db serialize.Schema
	err = json.Unmarshal(schemaJson, &db)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema json: %w", err)
	}

	return &db, nil
}

func UnmarshalJson(file *os.File) (*serialize.Schema, error) {
	bts, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var db serialize.Schema
	err = json.Unmarshal(bts, &db)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return &db, nil
}

// parseComments parses the comments from the file
// and returns the bytes of the file without the comments
func parseComments(file *os.File) ([]byte, error) {
	reader := bufio.NewReader(file)
	var result bytes.Buffer
	for {
		line, err := reader.ReadString('\n')

		if err != nil && err != io.EOF {
			fmt.Println("Error reading file:", err)
			return nil, err
		}

		line = removeComments(line)
		result.WriteString(line)

		if err == io.EOF {
			break
		}
	}

	return result.Bytes(), nil
}

// removeComments removes the comments from the line
func removeComments(line string) string {
	// Check if the line contains a comment
	if idx := strings.Index(line, "//"); idx != -1 {
		// Check if the comment is within a string (either single, double, or backtick quotes)
		quoteIdxDouble := strings.Index(line[:idx], "\"")
		quoteIdxSingle := strings.Index(line[:idx], "'")
		quoteIdxBacktick := strings.Index(line[:idx], "`")
		isInString := false

		if quoteIdxDouble != -1 && strings.Contains(line[quoteIdxDouble+1:], "'") {
			isInString = true
		}

		if quoteIdxSingle != -1 && strings.Contains(line[quoteIdxSingle+1:], "'") {
			isInString = true
		}

		if quoteIdxBacktick != -1 && strings.Contains(line[quoteIdxBacktick+1:], "'") {
			isInString = true
		}

		if !isInString {
			return line[:idx] + "\n"
		}
	}
	return line
}
