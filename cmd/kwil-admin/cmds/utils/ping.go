package utils

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var (
	pingLong = `Check connectivity with the node's admin RPC interface. If successful, returns 'pong'.`

	pingExample = `# Ping the node's admin RPC interface
kwil-admin node ping --rpcserver localhost:50151 --authrpc-cert "~/.kwild/rpc.cert"`
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "ping",
		Short:   "Check connectivity with the node's admin RPC interface.",
		Long:    pingLong,
		Example: pingExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			pong, err := client.Ping(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespString(pong))
		},
	}

	common.BindRPCFlags(cmd)

	return cmd
}
