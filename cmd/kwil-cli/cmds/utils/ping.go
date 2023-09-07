package utils

import (
	"context"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"

	"github.com/spf13/cobra"
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping is used to ping the kwil provider endpoint",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			var res string
			err := common.DialClient(cmd.Context(), common.WithoutPrivateKey, func(ctx context.Context, client *client.Client, cfg *config.KwilCliConfig) error {
				var _err error
				res, _err = client.Ping(ctx)
				return _err

			})

			msg := display.WrapMsg(respStr(res), err)
			return display.Print(msg, err, config.GetOutputFormat())
		},
	}

	return cmd
}
