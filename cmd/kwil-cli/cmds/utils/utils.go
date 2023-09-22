package utils

import (
	"github.com/spf13/cobra"
)

func NewCmdUtils() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "utils",
		Aliases: []string{"common"},
		Short:   "Various utility commands.",
		Long:    "",
	}

	cmd.AddCommand(
		signCmd(),
		pingCmd(),
		printConfigCmd(),
		walletCmd(),
		privateKeyCmd(), // TODO: use walletCmd instead, it's more intuitive
		txQueryCmd(),
	)

	return cmd
}
