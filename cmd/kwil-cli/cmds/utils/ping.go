package utils

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/spf13/cobra"
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping the kwil provider endpoint.  If successful, returns 'pong'.",
		Long:  "Ping the kwil provider endpoint.  If successful, returns 'pong'.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), cmd, common.WithoutPrivateKey, func(ctx context.Context, client common.Client, cfg *config.KwilCliConfig) error {
				res, err := client.Ping(ctx)
				if err != nil {
					return err
				}

				return display.PrintCmd(cmd, display.RespString(res))
			})
		},
	}

	return cmd
}
