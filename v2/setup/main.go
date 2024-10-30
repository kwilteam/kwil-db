package main

import (
	"os"

	"kwil/app"

	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	desc := "Setup is a tool to generate configuration for multiple nodes and generate keys for testing purposes"
	cmd := &cobra.Command{
		Use:               "setup",
		Short:             desc,
		Long:              desc,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDescriptions: true,
		},
		PersistentPreRunE: app.PreRunBindFlags,
	}

	// "root" does not have config file analog
	cmd.PersistentFlags().StringP(app.RootFlagName, "r", ".testnet", "root directory")

	cmd.AddCommand(app.TestnetCmd())
	cmd.AddCommand(app.KeygenCmd())
	cmd.AddCommand(app.ResetCmd())

	return cmd
}

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(-1)
	}
}
