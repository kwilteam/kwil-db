package database

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/common/display"
	grpc_client "kwil/kwil/client/grpc-client"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
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
		Use:   "execute",
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := grpc_client.NewClient(cc)
				if err != nil {
					return err
				}

				// check that args is odd and has at least 3 elements
				if len(args) < 3 || len(args)%2 == 0 {
					return fmt.Errorf("invalid number of arguments")
				}

				// if we get an error, it means the user did not specify the database id
				// get the database name and owner
				dbName, err := cmd.Flags().GetString("db_name")
				if err != nil {
					return fmt.Errorf("either database id or database name and owner must be specified: %w", err)
				}

				dbOwner, err := cmd.Flags().GetString("db_owner")
				if err != nil {
					return fmt.Errorf("either database id or database name and owner must be specified: %w", err)
				}

				res, err := client.ExecuteDatabase(ctx, dbOwner, dbName, args[0], args[1:])
				if err != nil {
					return fmt.Errorf("error executing database: %w", err)
				}

				// print the response
				display.PrintTxResponse(res)

				return nil
			})
		},
	}

	cmd.Flags().StringP("db_id", "i", "", "the database id")
	cmd.Flags().StringP("db_name", "n", "", "the database name")
	cmd.Flags().StringP("db_owner", "o", "", "the database owner")
	return cmd
}
