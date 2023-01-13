package fund

import (
	"kwil/cmd/kwil-cli/common"

	"github.com/spf13/cobra"
)

func NewCmdFund() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "fund",
		Short: "fund contains subcommands for funding",
		Long:  `With "fund" you can deposit, withdraw, and check your allowance.`,
	}

	common.BindKwilFlags(cmd)
	common.BindKwilEnv(cmd)

	common.BindChainFlags(cmd)
	common.BindChainEnv(cmd)

	cmd.AddCommand(
		approveCmd(),
		depositCmd(),
		withdrawCmd(),
		balancesCmd(),
	)

	return cmd
}
