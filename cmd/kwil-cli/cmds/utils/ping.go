package utils

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/spf13/cobra"
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping the kwil provider endpoint.  If successful, returns 'pong'.",
		Long:  "Ping the kwil provider endpoint.  If successful, returns 'pong'.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return common.DialClient(cmd.Context(), cmd, common.WithoutPrivateKey, func(ctx context.Context, client clientType.Client, cfg *config.KwilCliConfig) error {
				res, err := client.Ping(ctx)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				return display.PrintCmd(cmd, display.RespString(res))
			})
		},
	}

	return cmd
}
