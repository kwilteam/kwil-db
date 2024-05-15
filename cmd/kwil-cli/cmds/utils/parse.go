package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/internal/engine/ddl"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/spf13/cobra"
)

func newParseCmd() *cobra.Command {
	debug := false
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

			res, err := parse.ParseKuneiform(string(file))
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if !debug {
				res.Actions = nil
				res.Procedures = nil
			} else {
				debugSchema := "debug_schema"
				var tbls [][]string
				for _, t := range res.Schema.Tables {
					tbl, err := ddl.GenerateDDL(debugSchema, t)
					if err != nil {
						return display.PrintErr(cmd, err)
					}

					tbls = append(tbls, tbl)
				}

				var procs []string
				for _, p := range res.Procedures {
					proc, err := ddl.GenerateProcedure(p, debugSchema)
					if err != nil {
						return display.PrintErr(cmd, err)
					}

					procs = append(procs, proc)
				}

				var fprocs []string
				for _, p := range res.Schema.ForeignProcedures {
					proc, err := ddl.GenerateForeignProcedure(p)
					if err != nil {
						return display.PrintErr(cmd, err)
					}

					fprocs = append(fprocs, proc)
				}

				return display.PrintCmd(cmd, &dumpResult{
					ParseResult:                *res,
					GeneratedTables:            tbls,
					GeneratedProcedures:        procs,
					GeneratedForeignProcedures: fprocs,
				})
			}

			return display.PrintCmd(cmd, &schemaDisplay{Result: res})
		},
	}

	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "debug mode (show generated code and ASTs)")

	return cmd
}

// schemaDisplay is a struct that will be used to display the schema.
// It includes an error because the parser can return a schema even if there is an error.
type schemaDisplay struct {
	Result *parse.ParseResult
}

func (s *schemaDisplay) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Result)
}

func (s *schemaDisplay) MarshalText() (text []byte, err error) {
	// we set the schema info to nil because it is not needed for users
	// who are just looking at the schema.type ParseResult struct {
	s.Result.SchemaInfo = nil
	if err := s.Result.Err(); err != nil {
		return []byte(err.Error()), nil
	} else {
		return json.MarshalIndent(s.Result, "", "  ")
	}
}

type dumpResult struct {
	parse.ParseResult          `json:"base_result"`
	GeneratedTables            [][]string `json:"generated_tables"`
	GeneratedProcedures        []string   `json:"generated_procedures"`
	GeneratedForeignProcedures []string   `json:"generated_foreign_procedures"`
}

func (d *dumpResult) MarshalJSON() ([]byte, error) {
	return marshalSafe(d, json.Marshal)
}

func (d *dumpResult) MarshalText() (text []byte, err error) {
	return marshalSafe(d, func(v any) ([]byte, error) {
		return json.MarshalIndent(v, "", "  ")
	})
}

// marshalSafe will ignore custom MarshalJSON methods and marshal the struct as is.
// This is necessary to avoid infinite recursion when a struct has a custom MarshalJSON method.
func marshalSafe(v any, fn func(v any) ([]byte, error)) (bs []byte, err error) {
	k := reflect.TypeOf(v).Kind() // ptr or not?

	if k != reflect.Ptr {
		return fn(v)
	}

	// dereference pointer
	v2 := reflect.ValueOf(v).Elem().Interface()
	return marshalSafe(v2, fn)
}
