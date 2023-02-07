package database

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/internal/app/kcli/common/display"
	"kwil/internal/app/kcli/config"
	"kwil/pkg/kwil-client"
	"kwil/pkg/types/data_types/any_type"
)

const (
	ExecuteCmdLong = `Execute executes a query against the specified database.  The query name is
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

	create_user name satoshi age 32 --database-id x1234`
)

func executeCmd() *cobra.Command {
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
			clt, err := kwil_client.New(ctx, config.AppConfig)
			if err != nil {
				return err
			}

			// if we get an error, it means the user did not specify the database id
			// get the database name and owner
			dbName, err := cmd.Flags().GetString("db_name")
			if err != nil {
				return fmt.Errorf("either database id or database name and owner must be specified: %w", err)
			}

			inputs := make([]anytype.KwilAny, 0)
			for i := 1; i < len(args); i++ {
				in, err := anytype.New(args[i])
				if err != nil {
					return fmt.Errorf("error creating kwil any type with executable inputs: %w", err)
				}

				inputs = append(inputs, in)
			}

			res, err := clt.ExecuteDatabase(ctx, dbName, args[0], inputs)
			if err != nil {
				return fmt.Errorf("error executing database: %w", err)
			}

			// print the response
			display.PrintTxResponse(res)

			return nil
		},
	}

	cmd.Flags().StringP("db_name", "n", "", "the database name")
	return cmd
}
