package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/internal/app/kcli/config"
	"kwil/pkg/kclient"
)

func pingCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping is used to ping the kwil provider endpoint",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := kclient.New(ctx, config.AppConfig)
			if err != nil {
				return err
			}

			res, err := clt.Kwil.Ping(ctx)
			if err != nil {
				return fmt.Errorf("error pinging: %w", err)
			}
			fmt.Println(res)
			return nil
		},
	}

	return cmd
}
