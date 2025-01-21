package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/client"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/helpers"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/node/engine/parse"
	"github.com/spf13/cobra"
)

var (
	execSQLLong = `TODO: fill me out`

	execSQLExample = `TODO: fill me out
`
)

func execSQLCmd() *cobra.Command {
	var sqlStmt, sqlFilepath string
	var params []string

	cmd := &cobra.Command{
		Use:     "exec-sql",
		Short:   "Execute SQL against a database",
		Long:    execSQLLong,
		Example: execSQLExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txFlags, err := common.GetTxFlags(cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			params, err := parseParams(params)
			if err != nil {
				return err
			}

			var stmt string
			if len(args) > 0 {
				stmt = args[0]
			}
			if sqlStmt != "" {
				if stmt != "" {
					return display.PrintErr(cmd, fmt.Errorf(`received two SQL statements: "%s" and "%s"`, stmt, sqlStmt))
				}
				stmt = sqlStmt
			}
			if sqlFilepath != "" {
				if stmt != "" {
					return display.PrintErr(cmd, fmt.Errorf(`received two SQL statements: "%s" and file "%s"`, stmt, sqlFilepath))
				}

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

			if stmt == "" {
				return display.PrintErr(cmd, fmt.Errorf("no SQL statement provided"))
			}

			_, err = parse.Parse(stmt)
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to parse SQL statement: %s", err))
			}

			return client.DialClient(cmd.Context(), cmd, 0, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				txHash, err := cl.ExecuteSQL(ctx, stmt, params, clientType.WithNonce(txFlags.NonceOverride), clientType.WithSyncBroadcast(txFlags.SyncBroadcast))
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				return common.DisplayTxResult(ctx, cl, txHash, cmd)
			})
		},
	}

	cmd.Flags().StringVarP(&sqlStmt, "statement", "s", "", "the SQL statement to execute")
	cmd.Flags().StringVarP(&sqlFilepath, "file", "f", "", "the file containing the SQL statement(s) to execute")
	cmd.Flags().StringArrayVarP(&params, "param", "p", nil, `the parameters to pass to the SQL statement. format: "key:type=value"`)
	common.BindTxFlags(cmd)
	return cmd
}
