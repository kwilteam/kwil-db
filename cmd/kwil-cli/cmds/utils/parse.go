package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/spf13/cobra"
)

func newParseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse <file_path>",
		Short: "Parse a Kuneiform schema",
		Long:  `Parse a Kuneiform schema and output the JSON schema.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "" {
				return display.PrintErr(cmd, fmt.Errorf("file path is required"))
			}

			file, err := os.ReadFile(args[0])
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			res, err := parse.ParseAndValidate(file)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if res.Err() != nil {
				return display.PrintErr(cmd, res.Err())
			}

			return display.PrintCmd(cmd, &schemaDisplay{Result: res.Schema})
		},
	}

	return cmd
}

// schemaDisplay is a struct that will be used to display the schema.
type schemaDisplay struct {
	Result *types.Schema
}

func (s *schemaDisplay) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Result)
}

func (s *schemaDisplay) MarshalText() (text []byte, err error) {
	return json.MarshalIndent(s.Result, "", "  ")
}
