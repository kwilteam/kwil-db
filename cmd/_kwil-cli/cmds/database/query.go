package database

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
)

var (
	queryLong = `Query a database using an ad-hoc SQL SELECT statement.

Requires a SQL SELECT statement as an argument.

You can either specify the database to execute this against with the ` + "`" + `--name` + "`" + ` and ` + "`" + `--owner` + "`" + `
flags, or you can specify the database by passing the database id with the ` + "`" + `--dbid` + "`" + ` flag.  If a ` + "`" + `--name` + "`" + `
flag is passed and no ` + "`" + `--owner` + "`" + ` flag is passed, the owner will be inferred from your configured wallet.

Note that ad-hoc queries will be rejected on RPC servers that are operating with
authenticated call requests enabled.`

	queryExample = `# Querying the "users" table in the "mydb" database
kwil-cli database query "SELECT * FROM users WHERE age > 25" --name mydb --owner 0x9228624C3185FCBcf24c1c9dB76D8Bef5f5DAd64`
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     `query <select_statement>`,
		Short:   "Query a database using an ad-hoc SQL SELECT statement.",
		Long:    queryLong,
		Example: queryExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return helpers.DialClient(cmd.Context(), cmd, common.WithoutPrivateKey,
				func(ctx context.Context, client clientType.Client, conf *config.KwilCliConfig) error {
					dbid, err := getSelectedDbid(cmd, conf)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("target database not properly specified: %w", err))
					}

					data, err := client.Query(ctx, dbid, args[0])
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("error querying database: %w", err))
					}

					return display.PrintCmd(cmd, &respRelations{Data: data})
				})
		},
	}

	bindFlagsTargetingDatabase(cmd)
	return cmd
}
