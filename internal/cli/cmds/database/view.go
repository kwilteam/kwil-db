package database

import (
	"context"

	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
	"github.com/kwilteam/kwil-db/internal/cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func viewDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "view",
		Short: "View is used to view the details of a database.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, ksc v0.KwilServiceClient) error {
				resp, err := ksc.GetDatabase(ctx, &v0.GetDatabaseRequest{})
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

func listDatabaseCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List is used to list all databases.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return util.ConnectKwil(cmd.Context(), viper.GetViper(), func(ctx context.Context, ksc v0.KwilServiceClient) error {
				resp, err := ksc.ListDatabases(ctx, &v0.ListDatabasesRequest{})
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
