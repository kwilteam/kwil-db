package database

import (
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common/display"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"
	"kwil/pkg/databases/executables"
	"kwil/pkg/databases/spec"
	"strings"

	"github.com/spf13/cobra"
)

func executeCmd() *cobra.Command {
	var queryName string

	cmd := &cobra.Command{
		Use:   "execute [query field value [field value]...]",
		Short: "Execute a query",
		Long: `Execute executes a query against the specified database.  The query name is
specified as the first argument, and the query a arguments are specified after.
In order to specify an argument, you first need to specify the argument name.
You then specify the argument type.

For example, if I have a query name "create_user" that takes two arguments: name and age.
I would specify the query as follows:

create_user name satoshi age 32

You specify the database to execute this against with the --database-name flag, and
the owner with the --database-owner flag.

You can also specify the database by passing the database id with the --database-id flag.

For example:

create_user name satoshi age 32 --database-name mydb --database-owner 0xAfFDC06cF34aFD7D5801A13d48C92AD39609901D

OR

create_user name satoshi age 32 --database-id x1234`,
		Args: cobra.MatchAll(func(cmd *cobra.Command, args []string) error {
			// check that args is odd and has at least 3 elements
			if len(args) < 3 || len(args)%2 == 0 {
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

			lowerName := strings.ToLower(queryName)

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

	cmd.Flags().StringVarP(&queryName, queryNameFlag, "q", "", "the query name (required)")

	cmd.MarkFlagRequired(queryNameFlag)
	return cmd
}

func getInputs(executable *executables.QuerySignature, args []string) (map[string]*spec.KwilAny, error) {
	if len(args) < 2 || len(args)%2 != 0 {
		return nil, fmt.Errorf("invalid number of arguments")
	}

	stringInputs := make(map[string]string) // maps the arg name to the arg value
	for i := 0; i < len(args); i = i + 2 {
		stringInputs[strings.ToLower(args[i])] = args[i+1]
	}

	return executable.ConvertInputs(stringInputs)
}
