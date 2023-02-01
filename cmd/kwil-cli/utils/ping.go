package utils

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"kwil/cmd/kwil-cli/common"
	"kwil/internal/app/kcli"
	"kwil/pkg/fund"
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping is used to ping the kwil provider endpoint",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
				conf, err := fund.NewConfig()
				if err != nil {
					return fmt.Errorf("error getting client config: %w", err)
				}

				client, err := kcli.New(cc, conf)
				if err != nil {
					return fmt.Errorf("error creating client: %w", err)
				}

				res, err := client.Txs.Ping(ctx)
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
