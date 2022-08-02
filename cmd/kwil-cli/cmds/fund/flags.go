package fund

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/crypto"

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

	if conf.PrivateKey == nil {
		return "", fmt.Errorf("no private key found in config")
	}

	return crypto.AddressFromPrivateKey(conf.PrivateKey), nil
}
