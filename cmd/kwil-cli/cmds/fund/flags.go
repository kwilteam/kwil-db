package fund

import (
	"kwil/cmd/kwil-cli/conf"

	"github.com/spf13/cobra"
)

const (
	addressFlag = "address"
)

// getAddress returns the address flag value.
// If no value is passed, it will use the address of the user's wallet.
func getSelectedAddress(cmd *cobra.Command) (string, error) {
	if cmd.Flags().Changed(addressFlag) {
		return cmd.Flags().GetString(addressFlag)
	}

	return conf.GetWalletAddress()
}
