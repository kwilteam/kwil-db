package database

import (
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/types"
	"strings"

	"github.com/spf13/cobra"
)

func executeCmd() *cobra.Command {
	var actionName string

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a query",
		Long: `Execute executes a query against the specified database.  The query name is
specified as a required "--query" flag, and the query parameters as arguments.
In order to specify an parameter, you first need to specify the parameter name, then the parameter value.

For example, if I have a query name "create_user" that takes two arguments: name and age.
I would specify the query as follows:

name satoshi age 32 --query=create_user

You specify the database to execute this against with the --name flag, and
the owner with the --owner flag.

You can also specify the database by passing the database id with the --dbid flag.

For example:

create_user name satoshi age 32 --name mydb --owner 0xAfFDC06cF34aFD7D5801A13d48C92AD39609901D

OR

name satoshi age 32 --dbid=x1234 --query=create_user `,
		Args: cobra.MatchAll(func(cmd *cobra.Command, args []string) error {
			// check that args is even and has at least 2 elements
			if len(args) < 2 || len(args)%2 != 0 {
				return fmt.Errorf("invalid number of arguments")
			}
			return nil
		}),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := client.New(ctx, config.Config.Node.KwilProviderRpcUrl,
				client.WithoutServiceConfig(),
			)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			// if we get an error, it means the user did not specify the database id
			// get the database name and owner
			dbId, err := getSelectedDbid(cmd)
			if err != nil {
				return fmt.Errorf("target database not properly specified: %w", err)
			}

			lowerName := strings.ToLower(actionName)

			schema, err := clt.GetSchema(ctx, dbId)
			if err != nil {
				return fmt.Errorf("error getting schema: %w", err)
			}
			action := schema.GetAction(lowerName)
			if action == nil {
				return fmt.Errorf("action %s not found", lowerName)
			}

			executionBody := &models.ActionExecution{
				Action: lowerName,
				DBID:   dbId,
			}

			qry, err := clt.GetQuerySignature(ctx, dbId, lowerName)
			if err != nil {
				return fmt.Errorf("error getting query signature: %w", err)
			}

			inputs, err := getInputs(qry, args)
			if err != nil {
				return fmt.Errorf("error getting inputs: %w", err)
			}

			ecdsaPk, err := config.GetEcdsaPrivateKey()
			if err != nil {
				return fmt.Errorf("failed to get ecdsa key: %w", err)
			}

			res, err := clt.ExecuteDatabaseById(ctx, dbId, lowerName, inputs, ecdsaPk)
			if err != nil {
				return fmt.Errorf("error executing database: %w", err)
			}

			// print the response
			display.PrintTxResponse(res)

			return nil
		},
	}

	cmd.Flags().StringP(nameFlag, "n", "", "the database name")
	cmd.Flags().StringP(ownerFlag, "o", "", "the database owner")
	cmd.Flags().StringP(dbidFlag, "i", "", "the database id")

	cmd.Flags().StringVarP(&actionName, queryNameFlag, "q", "", "the query name (required)")

	cmd.MarkFlagRequired(queryNameFlag)
	return cmd
}

// inputs will be received as args.  The args will be in the form of
// $argname:value.  Example $username:satoshi $age:32
func getInputs(args []string) (map[string][]byte, error) {
	inputs := make(map[string][]byte)

	for _, arg := range args {
		if !strings.HasPrefix(arg, "$") {
			return nil, fmt.Errorf("invalid argument: %s.  argument must begin with '$'", arg)
		}

		// split the arg into name and value.  only split on the first ':'
		split := strings.SplitN(arg, ":", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid argument: %s.  argument must be in the form of $name:value", arg)
		}

		serialized, err := types.New()

	}
}
