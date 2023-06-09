package validator

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:     "validator",
		Aliases: []string{"val"},
		Short:   "manage validators",
		Long:    "Validator is a command that contains subcommands for handling the validators",
	}
)

func NewCmdValidator() *cobra.Command {
	rootCmd.AddCommand(
		approveCmd(),
	)

	return rootCmd
}
