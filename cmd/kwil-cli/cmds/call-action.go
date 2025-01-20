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
	"github.com/spf13/cobra"
)

var (
	callActionLong = `TODO: fill me out`

	callActionExample = `TODO: fill me out`
)

func callActionCmd() *cobra.Command {
	var namespace string
	var namedParams []string
	var gwAuth, rpcAuth, logs bool

	cmd := &cobra.Command{
		Use:     "call-action",
		Short:   "Call a view action",
		Long:    callActionLong,
		Example: callActionExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return display.PrintErr(cmd, fmt.Errorf("no action provided"))
			}

			tblConf, err := getTableConfig(cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// positional parameters
			var params []any
			for _, p := range args[1:] {
				_, param, err := parseTypedParam(p)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				params = append(params, param)
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

			return client.DialClient(cmd.Context(), cmd, dialFlags, func(ctx context.Context, cl clientType.Client, conf *config.KwilCliConfig) error {
				// if named params are specified, we need to query the action to find their positions
				if len(namedParams) > 0 {
					paramList, err := GetParamList(ctx, cl.Query, namespace, args[0])
					if err != nil {
						return display.PrintErr(cmd, err)
					}

					_, values, pos, err := getNamedParams(paramList, namedParams)
					if err != nil {
						return display.PrintErr(cmd, err)
					}
					// there is a case where an action has 3 parameters, but only 2 are specified positionally,
					// with the 3rd being specified as a named parameter. In this case, we need to ensure that the
					// length of params is the same as the length of actionParams
					for i, p := range pos {
						if p >= len(params) {
							params = append(params, make([]any, p-len(params)+1)...)
						}

						params[p] = values[i]
					}
				}

				res, err := cl.Call(ctx, namespace, args[0], params)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				return display.PrintCmd(cmd, &respCall{Data: res, PrintLogs: logs, tableConf: tblConf})
			})
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace to execute the action in")
	cmd.Flags().StringArrayVarP(&namedParams, "param", "p", nil, "named parameters that will override any positional parameters")
	cmd.Flags().BoolVar(&rpcAuth, "rpc-auth", false, "signals that the call is being made to a kwil node and should be authenticated with the private key")
	cmd.Flags().BoolVar(&gwAuth, "gateway-auth", false, "signals that the call is being made to a gateway and should be authenticated with the private key")
	cmd.Flags().BoolVar(&logs, "logs", false, "result will include logs from notices raised during the call")
	bindTableOutputFlags(cmd)

	return cmd
}

type respCall struct {
	Data      *types.CallResult
	PrintLogs bool
	tableConf *tableConfig
}

func (r *respCall) MarshalJSON() ([]byte, error) {
	if !r.PrintLogs {
		return json.Marshal(r.Data.QueryResult) // this is for backwards compatibility
	}

	bts, err := json.Marshal(r.Data)
	if err != nil {
		return nil, err
	}

	return bts, nil
}

func (r *respCall) MarshalText() (text []byte, err error) {
	if !r.PrintLogs {
		return recordsToTable(r.Data.QueryResult.ExportToStringMap(), r.tableConf), nil
	}

	bts := recordsToTable(r.Data.QueryResult.ExportToStringMap(), r.tableConf)

	if len(r.Data.Logs) > 0 {
		bts = append(bts, []byte("\n\nLogs:")...)
		for _, log := range r.Data.Logs {
			bts = append(bts, []byte("\n  "+log)...)
		}
	}

	return bts, nil
}
