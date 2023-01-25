package utils

import (
	"context"
	"fmt"
	"kwil/cmd/kwil-cli/common"
	grpc_client "kwil/kwil/client/grpc-client"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping is used to ping the kwil provider endpoint",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
				client, err := grpc_client.NewClient(cc)
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
