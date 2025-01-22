package cmds

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/client"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/engine/parse"
	"github.com/spf13/cobra"
)

var (
	queryLong = `Execute a SELECT statement against the database.

This command executes a SELECT statement against the database and formats the results in a table.
It can only be used to execute SELECT statements, and cannot be used with any other type of SQL statement.
If you need to execute a SQL statement that modifies the database, use the 'exec-sql' command.

It is not required to have a private key configured, unless the RPC you are calling is in private mode, or
you are talking to Kwil Gateway.`

	queryExample = `# Execute a simple SELECT statement
kwil-cli query "SELECT * FROM my_table"

# Execute a SELECT statement with a named parameter
kwil-cli query "SELECT * FROM my_table WHERE id = $id" --param id:int=1`
)

func queryCmd() *cobra.Command {
	var namedParams []string
	var gwAuth, rpcAuth bool

	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Execute a SELECT statement against the database",
		Long:    queryLong,
		Example: queryExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return display.PrintErr(cmd, fmt.Errorf("SELECT statement must be the only argument"))
			}

			tblConf, err := getTableConfig(cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			var dialFlags uint8
			if gwAuth {
				// if calling kgw, then not only do we need a private key, but we also need to authenticate
				dialFlags = client.UsingGateway
			}
			if rpcAuth {
				// if calling a kwil node, then we need to authenticate
				dialFlags = dialFlags | client.AuthenticatedCalls
			}
			if dialFlags == 0 {
				// if neither of the above, private key is not required
				dialFlags = client.WithoutPrivateKey
			}

			params, err := parseParams(namedParams)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			_, err = parse.Parse(args[0])
			if err != nil {
				return display.PrintErr(cmd, fmt.Errorf("failed to parse SQL statement: %s", err))
			}

			return client.DialClient(cmd.Context(), cmd, dialFlags, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				res, err := cl.Query(ctx, args[0], params)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				return display.PrintCmd(cmd, &respRelations{Data: res, conf: tblConf})
			})
		},
	}

	cmd.Flags().StringArrayVarP(&namedParams, "param", "p", nil, `named parameters that will be used in the query. format: "key:type=value"`)
	cmd.Flags().BoolVar(&rpcAuth, "rpc-auth", false, "signals that the call is being made to a kwil node and should be authenticated with the private key")
	cmd.Flags().BoolVar(&gwAuth, "gateway-auth", false, "signals that the call is being made to a gateway and should be authenticated with the private key")
	bindTableOutputFlags(cmd)
	return cmd
}

// respRelations is a slice of maps that represent the relations(from set theory)
// of a database in cli
type respRelations struct {
	// to avoid recursive call of MarshalJSON
	Data *types.QueryResult
	// conf for table formatting
	conf *tableConfig
}

func (r *respRelations) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Data)
}

func (r *respRelations) MarshalText() ([]byte, error) {
	return recordsToTable(r.Data.ExportToStringMap(), r.conf), nil
}
