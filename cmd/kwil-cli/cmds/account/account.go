package account

import (
	"github.com/spf13/cobra"
)

var nonceOverride int64

func NewCmdAccount() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "account",
		Short: "Account related commands.",
		Long:  "Commands related to Kwil account, such as balance checks and transfers.",
	}

	trCmd := transferCmd() // gets the nonce override flag

	cmd.AddCommand(
		balanceCmd(),
		trCmd,
	)

	trCmd.Flags().Int64VarP(&nonceOverride, "nonce", "N", -1, "nonce override (-1 means request from server)")

	return cmd
}
