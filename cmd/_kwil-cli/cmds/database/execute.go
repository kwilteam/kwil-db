package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
)

var (
	executeLong = `Execute a procedure or action against a database.

The procedure or action name is specified as the first positional argument, and the procedure parameters as all subsequent arguments.
In order to specify a procedure parameter, you first need to specify the parameter name, then the parameter value, delimited by a colon.

For example, for procedure ` + "`" + `get_user($username)` + "`" + `, you would specify the procedure as follows:
` + "`" + `execute get_user username:satoshi` + "`" + `

You can either specify the database to execute this against with the ` + "`" + `--name` + "`" + ` and ` + "`" + `--owner` + "`" + `
flags, or you can specify the database by passing the database id with the ` + "`" + `--dbid` + "`" + ` flag.  If a ` + "`" + `--name` + "`" + `
flag is passed and no ` + "`" + `--owner` + "`" + ` flag is passed, the owner will be inferred from your configured wallet.`

	executeExample = `# Executing the ` + "`" + `create_user($username, $age)` + "`" + ` procedure on the "mydb" database
kwil-cli database execute create_user username:satoshi age:32 --name mydb --owner 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64

# Executing the ` + "`" + `create_user($username, $age)` + "`" + ` procedure on a database using a dbid
kwil-cli database execute create_user username:satoshi age:32 --dbid 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64`
)

func executeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "execute <procedure_or_action> <parameter_1:value_1> <parameter_2:value_2> ...",
		Short:   "Execute a procedure or action against a database.",
		Long:    executeLong,
		Example: executeExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return helpers.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				dbid, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error getting selected dbid from CLI flags: %w", err))
				}

				action, args, err := getSelectedActionOrProcedure(cmd, args)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error getting selected action or procedure: %w", err))
				}

				parsedArgs, err := parseInputs(args)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error parsing inputs: %w", err))
				}

				inputs, err := buildExecutionInputs(ctx, cl, dbid, action, parsedArgs)
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("error getting inputs: %w", err))
				}

				// Could actually just directly pass nonce to the client method,
				// but those methods don't need tx details in the inputs.
				txHash, err := cl.Execute(ctx, dbid, action, inputs,
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

	bindFlagsTargetingProcedureOrAction(cmd)
	return cmd
}

// inputs will be received as args.  The args will be in the form of
// $argname:value.  Example $username:satoshi $age:32
func parseInputs(args []string) ([]map[string]string, error) {
	inputs := make(map[string]string, len(args))

	for _, arg := range args {
		ensureInputFormat(&arg)

		// split the arg into name and value.  only split on the first ':'
		split := strings.SplitN(arg, ":", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid argument: %s.  argument must be in the form of $name:value", arg)
		}

		inputs[split[0]] = split[1]
	}

	return []map[string]string{inputs}, nil
}
