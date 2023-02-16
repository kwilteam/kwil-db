package utils

import (
	"fmt"
	"kwil/cmd/kwil-cli/conf"
	"kwil/pkg/client"

	"github.com/spf13/cobra"
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping is used to ping the kwil provider endpoint",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := client.New(ctx, conf.Config.Node.KwilProviderRpcUrl,
				client.WithoutServiceConfig(),
			)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			res, err := clt.Ping(ctx)
			if err != nil {
				return fmt.Errorf("error pinging: %w", err)
			}
			fmt.Println(res)
			return nil
		},
	}

	return cmd
}
