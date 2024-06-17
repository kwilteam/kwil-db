package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/generate"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/spf13/cobra"
)

func newParseCmd() *cobra.Command {
	var debug, includePositions bool
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

			if includePositions && !debug {
				return display.PrintErr(cmd, fmt.Errorf("include-positions flag can only be used with debug"))
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
			// We also want to attempt to generate plpgsql functions.
			dis := &debugDisplay{
				Result:    res,
				Generated: generateAll(res.Schema),
			}

			if !includePositions {
				parse.RecursivelyVisitPositions(dis, func(gp parse.GetPositioner) {
					gp.Clear()
				})
			}

			if out == "" {
				return display.PrintCmd(cmd, dis)
			}

			bts, err := dis.MarshalText()
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
	cmd.Flags().BoolVarP(&includePositions, "include-positions", "p", false, "Include positions in the debug output")
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
	Result    *parse.SchemaParseResult `json:"parse_result"`
	Generated *genResult               `json:"generated"`
}

func (d *debugDisplay) MarshalJSON() ([]byte, error) {
	type res debugDisplay // prevent recursion
	return json.Marshal((*res)(d))
}

func (d *debugDisplay) MarshalText() (text []byte, err error) {
	return json.MarshalIndent(d, "", "  ")
}

// generateAll attempts to generate all ddl statements, sql, and plpgsql functions.
func generateAll(schema *types.Schema) *genResult {
	r := genResult{
		Tables:            make(map[string][]string),
		Actions:           make(map[string][]generate.GeneratedActionStmt),
		Procedures:        make(map[string]string),
		ForeignProcedures: make(map[string]string),
		Errors:            make([]error, 0),
	}
	defer func() {
		// catch any panics
		if e := recover(); e != nil {
			e2, ok := e.(error)
			if !ok {
				r.Errors = append(r.Errors, fmt.Errorf("panic: %v", e))
			} else {
				r.Errors = append(r.Errors, e2)
			}
		}
	}()

	wrapErr := func(s string, e error) error {
		return fmt.Errorf("%s: %w", s, e)
	}

	var err error
	for _, table := range schema.Tables {
		r.Tables[table.Name], err = generate.GenerateDDL(schema.Name, table)
		if err != nil {
			r.Errors = append(r.Errors, wrapErr("table "+table.Name, err))
		}
	}

	for _, action := range schema.Actions {
		r.Actions[action.Name], err = generate.GenerateActionBody(action, schema, schema.Name)
		if err != nil {
			r.Errors = append(r.Errors, wrapErr("action "+action.Name, err))
		}
	}

	for _, proc := range schema.Procedures {
		r.Procedures[proc.Name], err = generate.GenerateProcedure(proc, schema, schema.Name)
		if err != nil {
			r.Errors = append(r.Errors, wrapErr("procedure "+proc.Name, err))
		}
	}

	for _, proc := range schema.ForeignProcedures {
		r.ForeignProcedures[proc.Name], err = generate.GenerateForeignProcedure(proc, schema.Name, schema.DBID())
		if err != nil {
			r.Errors = append(r.Errors, wrapErr("foreign procedure "+proc.Name, err))
		}
	}

	return &r
}

type genResult struct {
	Tables            map[string][]string                       `json:"tables"`
	Actions           map[string][]generate.GeneratedActionStmt `json:"actions"`
	Procedures        map[string]string                         `json:"procedures"`
	ForeignProcedures map[string]string                         `json:"foreign_procedures"`
	Errors            []error                                   `json:"gen_errors"`
}
