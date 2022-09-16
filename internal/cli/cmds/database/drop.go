package database

import (
	"context"

	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func dropDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "drop",
		Short: "Drop is used to delete a database.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, ksc v0.KwilServiceClient) error {
				resp, err := ksc.DeleteDatabase(ctx, &v0.DeleteDatabaseRequest{})
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
