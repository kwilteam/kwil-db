package utils

import (
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/spf13/cobra"
)

func printConfigCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "print-config",
		Short: "Print the current CLI configuration.",
		Long:  "Print the current CLI configuration.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadCliConfig()
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &respKwilCliConfig{cfg: cfg})
		},
	}

	return cmd
}
