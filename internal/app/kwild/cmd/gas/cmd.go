package gas

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "gas-costs",
		Aliases: []string{"gas"},
		Short:   "manage gas costs",
		Long:    "gas-costs is a command that contains subcommands for handling the gas costs of the validators",
	}
)

func NewGasCmd() *cobra.Command {
	rootCmd.AddCommand(
		enableGasCmd(),
		disableGasCmd(),
	)

	return rootCmd
}
