package utils

import (
	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/spf13/cobra"
)

func printConfigCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "print-config",
		Short: "Print the current CLI configuration.",
		Long:  "Print the current CLI configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadCliConfig()
			if err != nil {
				return err
			}

			return display.PrintCmd(cmd, &respKwilCliConfig{cfg: cfg})
		},
	}

	return cmd
}
