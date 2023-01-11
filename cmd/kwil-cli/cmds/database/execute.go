package database

/*
func executeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
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
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := client.NewClient(cc, viper.GetViper())
				if err != nil {
					return fmt.Errorf("failed to create client: %w", err)
				}

				// check that args is odd and has at least 3 elements
				if len(args) < 3 || len(args)%2 == 0 {
					return fmt.Errorf("invalid number of arguments")
				}

				// we will check if the user specified the database id or the database name and owner
				isDbId := false
				var dbName string
				var dbOwner string

				// get the database id
				dbId, err := cmd.Flags().GetString("database-id")
				if err == nil {
					isDbId = true
				} else {

				}

				// get the database name and owner
				dbName, err := cmd.Flags().GetString("database-name")
				if err != nil {
					return fmt.Errorf("must specify a database name with the --database-name flag")
				}
			})
		},
	}
	return cmd
}
*/
