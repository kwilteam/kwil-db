package database

import (
	"context"
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
	var authenticate bool

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
					data = &clientType.Records{}
				}

				return display.PrintCmd(cmd, &respRelations{Data: data})
			})
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the target database schema name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the target database schema owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the target database id")
	cmd.Flags().StringVarP(&action, actionNameFlag, "a", "", "the target action name (required)")
	cmd.Flags().BoolVar(&authenticate, "authenticate", false, "authenticate signals that the call is being made to a gateway and should be authenticated with the private key")

	cmd.MarkFlagRequired(actionNameFlag)
	return cmd
}

// buildProcedureInputs will build the inputs for either
// an action or procedure executon/call.
func buildExecutionInputs(ctx context.Context, client clientType.Client, dbid string, proc string, inputs []map[string]any) ([][]any, error) {
	schema, err := client.GetSchema(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf("error getting schema: %w", err)
	}

	for _, a := range schema.Actions {
		if strings.EqualFold(a.Name, proc) {
			return buildActionInputs(a, inputs), nil
		}
	}

	for _, p := range schema.Procedures {
		if strings.EqualFold(p.Name, proc) {
			return buildProcedureInputs(p, inputs), nil
		}
	}

	return nil, fmt.Errorf("procedure/action not found")
}

func buildActionInputs(a *types.Action, inputs []map[string]any) [][]any {
	tuples := [][]any{}
	for _, input := range inputs {
		newTuple := []any{}
		for _, inputField := range a.Parameters {
			newTuple = append(newTuple, input[inputField])
		}

		tuples = append(tuples, newTuple)
	}

	return tuples
}

func buildProcedureInputs(p *types.Procedure, inputs []map[string]any) [][]any {
	tuples := [][]any{}
	for _, input := range inputs {
		newTuple := []any{}
		for _, inputField := range p.Parameters {
			newTuple = append(newTuple, input[inputField.Name])
		}

		tuples = append(tuples, newTuple)
	}

	return tuples
}
