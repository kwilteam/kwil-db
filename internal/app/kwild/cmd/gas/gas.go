package gas

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/spf13/cobra"
)

// ApproveCmd is used for approving validators
func enableGasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enables gas prices on the transactions",
		Long:  "The enable command is used to enable the gas costs on all the transactions on the validator nodes.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg, err := config.LoadKwildConfig()
			if err != nil {
				return err
			}
			options := []client.ClientOpt{}

			clt, err := client.New(ctx, cfg.GrpcListenAddress, options...)
			if err != nil {
				return err
			}

			err = clt.UpdateGasCosts(ctx, true)
			return err
		},
	}
	return cmd
}

func disableGasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disables gas prices on the transactions",
		Long:  "The disable command is used to disable the gas costs on all the transactions on the validator node.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg, err := config.LoadKwildConfig()
			if err != nil {
				return err
			}
			options := []client.ClientOpt{}

			clt, err := client.New(ctx, cfg.GrpcListenAddress, options...)
			if err != nil {
				return err
			}

			err = clt.UpdateGasCosts(ctx, false)
			return err
		},
	}
	return cmd
}
