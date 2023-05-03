package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"strings"

	"github.com/spf13/cobra"
)

func executeCmd() *cobra.Command {
	var actionName string

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a query",
		Long: `Execute executes a query against the specified database.  The query name is
specified as a required "--action" flag, and the query parameters as arguments.
In order to specify an parameter, you first need to specify the parameter name, then the parameter value, delimited by a colon.
You can include the input's '$' prefix if you wish, but it is not required.

For example, if I have a query name "create_user" that takes two arguments: name and age.
I would specify the query as follows:

'$name:satoshi' '$age:32' --action=create_user

You specify the database to execute this against with the --name flag, and
the owner with the --owner flag.

You can also specify the database by passing the database id with the --dbid flag.

For example:

'$name:satoshi' 'age:32' --action=create_user --name mydb --owner 0xAfFDC06cF34aFD7D5801A13d48C92AD39609901D

OR

'$name:satoshi' '$age:32' --dbid=x1234 --action=create_user `,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error {
				dbId, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return fmt.Errorf("target database not properly specified: %w", err)
				}

				lowerName := strings.ToLower(actionName)

				inputs, err := getInputs(args)
				if err != nil {
					return fmt.Errorf("error getting inputs: %w", err)
				}

				receipt, results, err := client.ExecuteAction(ctx, dbId, lowerName, inputs)
				if err != nil {
					return fmt.Errorf("error executing database: %w", err)
				}

				// print the response
				display.PrintTxResponse(receipt)

				// print the results
				printTableAny(results)

				return nil
			})
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the database id")

	cmd.Flags().StringVarP(&actionName, actionNameFlag, "a", "", "the action name (required)")

	cmd.MarkFlagRequired(actionNameFlag)
	return cmd
}

// inputs will be received as args.  The args will be in the form of
// $argname:value.  Example $username:satoshi $age:32
func getInputs(args []string) ([]map[string]any, error) {
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

func printActionResults(results []map[string]any) {
	for _, row := range results {
		for k, v := range row {
			fmt.Printf("%s: %v", k, v)
		}
	}
}
