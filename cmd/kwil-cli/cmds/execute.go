package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/client"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
)

var (
	executeLong = `Execute SQL or an action against a database.

To execute a SQL statement against the database, use the --sql flag.  The SQL statement will be executed with the
arguments passed as parameters.  The arguments are specified as $name:value.  For example, to execute the SQL
statement "SELECT * FROM users WHERE age > 25", you would specify the following:
` + "`" + `kwil-cli database execute --sql "INSERT INTO ids (id) VALUES ($age);" age:25` + "`" + `

To specify an action to execute, you can pass the action name as the first positional argument, or as the --action flag.
The action name is specified as the first positional argument, and the action parameters as all subsequent arguments.`

	executeExample = `# Executing a CREATE TABLE statement on the "mydb" database
kwil-cli database execute --sql "CREATE TABLE users (id UUID, name TEXT, age INT8);" --namespace mydb
	
# Executing the ` + "`" + `create_user($username, $age)` + "`" + ` procedure on the "mydb" database
kwil-cli database execute --action create_user username:satoshi age:32 --namespace mydb
`
)

func execSQLCmd() *cobra.Command {
	var sqlStmt, sqlFilepath string
	var params []string

	cmd := &cobra.Command{
		Use:     "exec-sql",
		Short:   "Execute SQL against a database",
		Long:    executeLong,
		Example: executeExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := parseParams(params)
			if err != nil {
				return err
			}

			if sqlStmt == "" && sqlFilepath == "" {
				return display.PrintErr(cmd, fmt.Errorf("no SQL statement provided"))
			}
			if sqlStmt != "" && sqlFilepath != "" {
				return display.PrintErr(cmd, fmt.Errorf("cannot provide both a SQL statement and a file"))
			}

			if sqlFilepath != "" {
				expanded, err := helpers.ExpandPath(sqlFilepath)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				file, err := os.ReadFile(expanded)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				sqlStmt = string(file)
			}

			return client.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				cl.ExecuteSQL(ctx, sqlStmt, params)
			})
		},
	}

	cmd.Flags().StringVarP(&sqlStmt, "statement", "s", "", "the SQL statement to execute")
	cmd.Flags().StringVarP(&sqlFilepath, "file", "f", "", "the file containing the SQL statement(s) to execute")
	cmd.Flags().StringArrayVarP(&params, "param", "p", nil, "the parameters to pass to the SQL statement")
	return cmd
}
