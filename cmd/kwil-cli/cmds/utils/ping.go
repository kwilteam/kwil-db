package utils

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/client"

	"github.com/spf13/cobra"
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping is used to ping the kwil provider endpoint",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialClient(cmd.Context(), common.WithoutServiceConfig, func(ctx context.Context, client *client.Client, config *config.KwilCliConfig) error {
				res, err := client.Ping(ctx)
				if err != nil {
					return fmt.Errorf("error pinging: %w", err)
				}
				fmt.Println(res)
				return nil
			})
		},
	}

	return cmd
}
