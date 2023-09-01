package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

func callCmd() *cobra.Command {
	var action string
	authenticate := new(bool)

	cmd := &cobra.Command{
		Use:   "call",
		Short: "Call an 'view' action",
		Long: `call an 'view' action that is a read-only action.
The query name is specified as a required "--action" flag, and the query parameters as arguments.
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
			return common.DialClient(cmd.Context(), 0, func(ctx context.Context, clnt *client.Client, conf *config.KwilCliConfig) error {
				dbid, err := getSelectedDbid(cmd, conf)
				if err != nil {
					return fmt.Errorf("target database not properly specified: %w", err)
				}

				lowerName := strings.ToLower(action)

				inputs, err := parseInputs(args)
				if err != nil {
					return fmt.Errorf("error getting inputs: %w", err)
				}

				actionStructure, err := getAction(ctx, clnt, dbid, lowerName)
				if err != nil {
					return fmt.Errorf("error getting action: %w", err)
				}

				tuples, err := createActionInputs(inputs, actionStructure)
				if err != nil {
					return fmt.Errorf("error creating action inputs: %w", err)
				}

				if len(tuples) == 0 {
					tuples = append(tuples, []any{})
				}

				res, err := clnt.CallAction(ctx, dbid, lowerName, tuples[0], client.Authenticated(*authenticate))
				if err != nil {
					return fmt.Errorf("error executing action: %w", err)
				}

				results := make([]map[string]string, len(res))
				for i, r := range res {
					results[i] = make(map[string]string)
					for k, v := range r {
						results[i][k] = fmt.Sprintf("%v", v)
					}
				}

				printTable(results)
				return nil
			})
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the database id")
	cmd.Flags().StringVarP(&action, actionNameFlag, "a", "", "the action name (required)")
	authenticate = cmd.Flags().BoolP("authenticate", "a", false, "whether to authenticate the action call")

	cmd.MarkFlagRequired(actionNameFlag)
	return cmd
}
