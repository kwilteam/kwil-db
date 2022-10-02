package database

import (
	"context"

	"kwil/x/cli/util"
	apipb "kwil/x/proto/apisvc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func dropDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "drop",
		Short: "Drop is used to delete a database.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, ksc apipb.KwilServiceClient) error {
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
