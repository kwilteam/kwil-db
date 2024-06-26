package utils

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/spf13/cobra"
)

var (
	chainInfoLong = `Display information about the connected Kwil network.`
)

func chainInfoCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "chain-info",
		Short: chainInfoLong,
		Long:  chainInfoLong,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return common.DialClient(cmd.Context(), cmd, common.WithoutPrivateKey, func(ctx context.Context, client clientType.Client, cfg *config.KwilCliConfig) error {
				chainInfo, err := client.ChainInfo(ctx)
				if err != nil {
					return display.PrintErr(cmd, err)
				}

				return display.PrintCmd(cmd, &respChainInfo{Info: chainInfo})
			})
		},
	}

	return cmd
}
