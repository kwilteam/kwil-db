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
	var debug bool
	var out string

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

			// if not in debug mode, throw any errors and swallow the schema,
			// since the schema is invalid.
			if !debug {
				if res.Err() != nil {
					return display.PrintErr(cmd, res.Err())
				}

				if out == "" {
					return display.PrintCmd(cmd, &schemaDisplay{Result: res.Schema})
				}

				bts, err := json.MarshalIndent(res.Schema, "", "  ")
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				err = os.WriteFile(out, bts, 0644)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				return display.PrintCmd(cmd, display.RespString(fmt.Sprintf("Schema written to %s", out)))
			}

			// if in debug mode, output the schema and the debug information.
			if out == "" {
				return display.PrintCmd(cmd, &debugDisplay{Result: res})
			}

			bts, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			err = os.WriteFile(out, bts, 0644)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespString(fmt.Sprintf("Debug information written to %s", out)))
		},
	}

	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Display debug information")
	cmd.Flags().StringVarP(&out, "out", "o", "", "Output file. If debug is true, errors will also be written to this file")

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

// debugDisplay is a struct that will be used to display the schema.
// It is used to display the debug information.
type debugDisplay struct {
	Result *parse.SchemaParseResult
}

func (d *debugDisplay) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Result)
}

func (d *debugDisplay) MarshalText() (text []byte, err error) {
	return json.MarshalIndent(d.Result, "", "  ")
}
