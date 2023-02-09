package database

import (
	"github.com/spf13/cobra"
	"kwil/internal/app/kcli/common/display"
	"kwil/internal/app/kcli/config"
	"kwil/pkg/kclient"
)

func dropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop db_name",
		Short: "Drops a database",
		Long:  "Drops a database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := kclient.New(ctx, config.AppConfig)
			if err != nil {
				return err
			}

			res, err := clt.DropDatabase(ctx, args[0])
			if err != nil {
				return err
			}

			display.PrintTxResponse(res)

			return nil
		},
	}
	return cmd
}
