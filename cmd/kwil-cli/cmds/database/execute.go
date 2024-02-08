package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/spf13/cobra"
)

var (
	executeLong = `Execute an action against a database.

The action name is specified as a required "--action" flag, and the action parameters as arguments.
In order to specify an action parameter, you first need to specify the parameter name, then the parameter value, delimited by a colon.

For example, for action ` + "`" + `get_user($username)` + "`" + `, you would specify the action as follows:
` + "`" + `username:satoshi` + "`" + ` --action=get_user

You can either specify the database to execute this against with the ` + "`" + `--name` + "`" + ` and ` + "`" + `--owner` + "`" + `
flags, or you can specify the database by passing the database id with the ` + "`" + `--dbid` + "`" + ` flag.  If a ` + "`" + `--name` + "`" + `
flag is passed and no ` + "`" + `--owner` + "`" + ` flag is passed, the owner will be inferred from your configured wallet.`

	executeExample = `# Executing the ` + "`" + `create_user($username, $age)` + "`" + ` action on the "mydb" database
kwil-cli database execute username:satoshi age:32 --action create_user --name mydb --owner 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64

# Executing the ` + "`" + `create_user($username, $age)` + "`" + ` action on a database using a dbid
kwil-cli database execute username:satoshi age:32 --action create_user --dbid 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64`
)

func executeCmd() *cobra.Command {
	var actionName string

	cmd := &cobra.Command{
		Use:     "execute <parameter_1:value_1> <parameter_2:value_2> ...",
		Short:   "Execute an action against a database.",
		Long:    executeLong,
		Example: executeExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				dbId, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("target database not properly specified: %w", err))
				}

				lowerName := strings.ToLower(actionName)

				actionStructure, err := getAction(ctx, cl, dbId, lowerName)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error getting action: %w", err))
				}

				inputs, err := GetInputs(args, actionStructure)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error getting inputs: %w", err))
				}

				// Could actually just directly pass nonce to the client method,
				// but those methods don't need tx details in the inputs.
				txHash, err := cl.ExecuteAction(ctx, dbId, lowerName, inputs,
					clientType.WithNonce(nonceOverride), clientType.WithSyncBroadcast(syncBcast))
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error executing database: %w", err))
				}
				// If sycnBcast, and we have a txHash (error or not), do a query-tx.
				if len(txHash) != 0 && syncBcast {
					time.Sleep(500 * time.Millisecond) // otherwise it says not found at first
					resp, err := cl.TxQuery(ctx, txHash)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("tx query failed: %w", err))
					}
					return display.PrintCmd(cmd, display.NewTxHashAndExecResponse(resp))
				}
				return display.PrintCmd(cmd, display.RespTxHash(txHash))
			})
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the target database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the target database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the target database id")

	cmd.Flags().StringVarP(&actionName, actionNameFlag, "a", "", "the target action name (required)")

	cmd.MarkFlagRequired(actionNameFlag)
	return cmd
}

// inputs will be received as args.  The args will be in the form of
// $argname:value.  Example $username:satoshi $age:32
func parseInputs(args []string) ([]map[string]any, error) {
	inputs := make(map[string]any)

	for _, arg := range args {
		ensureInputFormat(&arg)

		// split the arg into name and value.  only split on the first ':'
		split := strings.SplitN(arg, ":", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid argument: %s.  argument must be in the form of $name:value", arg)
		}

		inputs[split[0]] = split[1]
	}

	return []map[string]any{inputs}, nil
}

func GetInputs(args []string, action *transactions.Action) ([][]any, error) {
	inputs, err := parseInputs(args)
	if err != nil {
		return nil, fmt.Errorf("error getting inputs: %w", err)
	}

	return createActionInputs(inputs, action)
}

// createActionInputs takes a []map[string]any and an action, and converts it to [][]any
func createActionInputs(inputs []map[string]any, action *transactions.Action) ([][]any, error) {
	tuples := [][]any{}
	for _, input := range inputs {
		newTuple := []any{}
		for _, inputField := range action.Inputs {
			value, ok := input[inputField]
			if !ok {
				return nil, fmt.Errorf("missing input: %s", inputField)
			}

			newTuple = append(newTuple, value)
		}

		tuples = append(tuples, newTuple)
	}

	return tuples, nil
}
