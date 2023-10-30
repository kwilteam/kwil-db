package utils

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/spf13/cobra"
)

func chainInfoCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "chain-info",
		Short: "Ping is used to ping the kwil provider endpoint",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			var chainInfo *types.ChainInfo
			err := common.DialClient(cmd.Context(), common.WithoutPrivateKey, func(ctx context.Context, client *client.Client, cfg *config.KwilCliConfig) error {
				var err error
				chainInfo, err = client.ChainInfo(ctx)
				return err
			})

			return display.Print(&respChainInfo{chainInfo}, err, config.GetOutputFormat())
		},
	}

	return cmd
}
