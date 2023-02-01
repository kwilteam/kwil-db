package database

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"kwil/cmd/kwil-cli/common"
	"kwil/cmd/kwil-cli/common/display"
	"kwil/internal/app/kcli"
	"kwil/pkg/fund"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Drops a database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.DialGrpc(cmd.Context(), func(ctx context.Context, cc *grpc.ClientConn) error {
				conf, err := fund.NewConfig()
				if err != nil {
					return fmt.Errorf("error getting client config: %w", err)
				}

				client, err := kcli.New(cc, conf)
				if err != nil {
					return err
				}
				// should be one arg
				if len(args) != 1 {
					return fmt.Errorf("deploy requires one argument: database name")
				}

				res, err := client.DropDatabase(ctx, client.Fund.GetConfig().GetAccountAddress(), args[0])
				if err != nil {
					return err
				}

				display.PrintTxResponse(res)

				return nil
			})
		},
	}
	return cmd
}
