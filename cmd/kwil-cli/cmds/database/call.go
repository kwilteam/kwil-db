package database

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"

	"github.com/spf13/cobra"
)

var (
	callLong = `Call a ` + "`" + `view` + "`" + ` action, returning the result.

` + "`" + `view` + "`" + ` actions are read-only actions that do not require gas to execute.  They are
the primary way to query the state of a database. The ` + "`" + `call` + "`" + ` command is used to call
a ` + "`" + `view` + "`" + ` action on a database.  It takes the action name as a required flag, and the
action inputs as arguments.

To specify an action input, you first need to specify the input name, then the input value, delimited by a colon.
For example, for action ` + "`" + `get_user($username)` + "`" + `, you would specify the action as follows:

` + "`" + `username:satoshi` + "`" + ` --action=get_user

You can either specify the database to execute this against with the ` + "`" + `--name` + "`" + ` and ` + "`" + `--owner` + "`" + `
flags, or you can specify the database by passing the database id with the ` + "`" + `--dbid` + "`" + ` flag.  If a ` + "`" + `--name` + "`" + `
flag is passed and no ` + "`" + `--owner` + "`" + ` flag is passed, the owner will be inferred from your configured wallet.

If you are interacting with a Kwil gateway, you can also pass the ` + "`" + `--authenticate` + "`" + ` flag to authenticate the call with your private key.`

	callExample = `# Calling the ` + "`" + `get_user($username)` + "`" + ` action on the "mydb" database
kwil-cli database call --action get_user --name mydb --owner 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64 username:satoshi

# Calling the ` + "`" + `get_user($username)` + "`" + ` action on a database using a dbid, authenticating with a private key
kwil-cli database call --action get_user --dbid 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64 username:satoshi --authenticate`
)

func callCmd() *cobra.Command {
	var action string
	var authenticate, logs bool

	cmd := &cobra.Command{
		Use:     "call <parameter_1:value_1> <parameter_2:value_2> ...",
		Short:   "Call a 'view' action, returning the result.",
		Long:    callLong,
		Example: callExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			dialFlags := common.WithoutPrivateKey
			if authenticate {
				// overwrite the WithoutPrivateKey flag, and add the UsingGateway flag
				dialFlags = common.UsingGateway
			}

			return common.DialClient(cmd.Context(), cmd, dialFlags, func(ctx context.Context, clnt clientType.Client, conf *config.KwilCliConfig) error {
				dbid, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("target database not properly specified: %w", err))
				}

				lowerName := strings.ToLower(action)

				inputs, err := parseInputs(args)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error getting inputs: %w", err))
				}

				tuples, err := buildExecutionInputs(ctx, clnt, dbid, lowerName, inputs)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error creating action inputs: %w", err))
				}

				if len(tuples) == 0 {
					tuples = append(tuples, []any{})
				}

				data, err := clnt.Call(ctx, dbid, lowerName, tuples[0])
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error calling action: %w", err))
				}

				if data == nil {
					data = &clientType.CallResult{}
				}

				return display.PrintCmd(cmd, &respCall{
					Data:      data,
					PrintLogs: logs,
				})
			})
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the target database schema name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the target database schema owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the target database id")
	cmd.Flags().StringVarP(&action, actionNameFlag, "a", "", "the target action name (required)")
	cmd.Flags().BoolVar(&authenticate, "authenticate", false, "authenticate signals that the call is being made to a gateway and should be authenticated with the private key")
	cmd.Flags().BoolVar(&logs, "logs", false, "result will include logs from notices raised during the call")

	cmd.MarkFlagRequired(actionNameFlag)
	return cmd
}

type respCall struct {
	Data      *clientType.CallResult
	PrintLogs bool
}

func (r *respCall) MarshalJSON() ([]byte, error) {
	if !r.PrintLogs {
		return json.Marshal(r.Data.Records.ExportString()) // this is for backwards compatibility
	}

	bts, err := json.Marshal(r.Data)
	if err != nil {
		return nil, err
	}

	return bts, nil
}

func (r *respCall) MarshalText() (text []byte, err error) {
	if !r.PrintLogs {
		return recordsToTable(r.Data.Records), nil
	}

	bts := recordsToTable(r.Data.Records)

	if len(r.Data.Logs) > 0 {
		bts = append(bts, []byte("\n\nLogs:")...)
		for _, log := range r.Data.Logs {
			bts = append(bts, []byte("\n  "+log)...)
		}
	}

	return bts, nil
}

// buildProcedureInputs will build the inputs for either
// an action or procedure executon/call.
func buildExecutionInputs(ctx context.Context, client clientType.Client, dbid string, proc string, inputs []map[string]string) ([][]any, error) {
	schema, err := client.GetSchema(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf("error getting schema: %w", err)
	}

	for _, a := range schema.Actions {
		if strings.EqualFold(a.Name, proc) {
			return buildActionInputs(a, inputs)
		}
	}

	for _, p := range schema.Procedures {
		if strings.EqualFold(p.Name, proc) {
			return buildProcedureInputs(p, inputs)
		}
	}

	return nil, fmt.Errorf("procedure/action not found")
}

// decodeMany attempts to parse command-line inputs as base64 encoded values.
func decodeMany(inputs []string) ([][]byte, bool) {
	b64Arr := [][]byte{}
	b64Ok := true
	for _, s := range inputs {
		// in the CLI, if data has suffix ;b64, it is base64 encoded
		if strings.HasSuffix(s, "#b64") {
			s = strings.TrimSuffix(s, "#b64")
		} else {
			b64Ok = false
			break
		}

		bts, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			b64Ok = false
			break
		}
		b64Arr = append(b64Arr, bts)
	}

	return b64Arr, b64Ok
}

func buildActionInputs(a *types.Action, inputs []map[string]string) ([][]any, error) {
	tuples := [][]any{}
	for _, input := range inputs {
		newTuple := []any{}
		for _, inputField := range a.Parameters {
			// unlike procedures, actions do not have typed parameters,
			// so we should try to always parse arrays.

			val, ok := input[inputField]
			if !ok {
				fmt.Println(len(newTuple))
				// if not found, we should just add nil
				newTuple = append(newTuple, nil)
				continue
			}

			split, err := splitIgnoringQuotedCommas(val)
			if err != nil {
				return nil, err
			}

			// attempt to decode base64 encoded values
			b64Arr, b64Ok := decodeMany(split)
			if b64Ok {
				// additional check here in case user is sending a single base64 value, we don't
				// want to encode it as an array.
				if len(b64Arr) == 1 {
					newTuple = append(newTuple, b64Arr[0])
					continue
				}

				newTuple = append(newTuple, b64Arr)
			} else {
				// if nothing was split, then keep the original value, not the []string{}
				if len(split) == 1 {
					newTuple = append(newTuple, split[0])
					continue
				}

				newTuple = append(newTuple, split)
			}
		}
		tuples = append(tuples, newTuple)
	}

	return tuples, nil
}

func buildProcedureInputs(p *types.Procedure, inputs []map[string]string) ([][]any, error) {
	tuples := [][]any{}
	for _, input := range inputs {
		newTuple := []any{}
		for _, inputField := range p.Parameters {
			v, ok := input[inputField.Name]
			if !ok {
				// if not found, we should just add nil
				newTuple = append(newTuple, nil)
				continue
			}

			// if the input is an array, split it by commas
			if inputField.Type.IsArray {
				split, err := splitIgnoringQuotedCommas(v)
				if err != nil {
					return nil, err
				}

				// attempt to decode base64 encoded values
				b64Arr, b64Ok := decodeMany(split)
				if b64Ok {
					newTuple = append(newTuple, b64Arr)
				} else {
					newTuple = append(newTuple, split)
				}
				continue
			}

			// attempt to decode base64 encoded values

			bts, ok := decodeMany([]string{v})
			if ok {
				newTuple = append(newTuple, bts[0])
			} else {
				newTuple = append(newTuple, input[inputField.Name])
			}
		}

		tuples = append(tuples, newTuple)
	}

	return tuples, nil
}

// splitIgnoringQuotedCommas splits a string by commas, but ignores commas that are inside single or double quotes.
// It will return an error if there are unclosed quotes.
func splitIgnoringQuotedCommas(input string) ([]string, error) {
	var result []string
	var currentToken []rune
	inSingleQuote := false
	inDoubleQuote := false

	for _, char := range input {
		switch char {
		case '\'':
			if !inDoubleQuote { // Toggle single quote flag if not inside double quotes
				inSingleQuote = !inSingleQuote
				continue // Skip appending this quote character to token
			}
			currentToken = append(currentToken, char)
		case '"':
			if !inSingleQuote { // Toggle double quote flag if not inside single quotes
				inDoubleQuote = !inDoubleQuote
				continue // Skip appending this quote character to token
			}
			currentToken = append(currentToken, char)
		case ',':
			if inSingleQuote || inDoubleQuote { // If inside quotes, treat comma as a normal character
				currentToken = append(currentToken, char)
			} else { // Otherwise, it's a delimiter
				result = append(result, string(currentToken))
				currentToken = []rune{}
			}
		default:
			currentToken = append(currentToken, char)
		}
	}

	// Append the last token
	result = append(result, string(currentToken))

	if inSingleQuote || inDoubleQuote {
		return nil, fmt.Errorf("unclosed quote in array inputs")
	}

	return result, nil
}
