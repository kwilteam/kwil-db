package fund

import (
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/crypto"

	"github.com/spf13/cobra"
)

const (
	addressFlag = "address"
)

// getAddress returns the address flag value.
// If no value is passed, it will use the address of the user's wallet.
func getSelectedAddress(cmd *cobra.Command, conf *config.KwilCliConfig) (string, error) {
	if cmd.Flags().Changed(addressFlag) {
		return cmd.Flags().GetString(addressFlag)
	}

	return crypto.AddressFromPrivateKey(conf.PrivateKey), nil
}
