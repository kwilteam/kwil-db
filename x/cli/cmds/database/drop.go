package database

import (
	"context"

	"kwil/x/cli/util"
	"kwil/x/proto/apipb"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func dropDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "drop",
		Short: "Drop is used to delete a database.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, cc *grpc.ClientConn) error {
				ksc := apipb.NewKwilServiceClient(cc)
				resp, err := ksc.DeleteDatabase(ctx, &apipb.DeleteDatabaseRequest{})
				if err != nil {
					return err
				}
				_ = resp
				return nil
			})
		},
	}

	return cmd
}
