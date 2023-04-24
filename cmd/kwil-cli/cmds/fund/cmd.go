package fund

import (
	"github.com/spf13/cobra"
)

func NewCmdFund() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "fund",
		Short: "fund contains subcommands for funding",
		Long:  `With "fund" you can deposit, withdraw, and check your allowance.`,
	}

	cmd.AddCommand(
		approveCmd(),
		depositCmd(),
		balancesCmd(),
		getAccountCmd(),
	)
	return cmd
}
