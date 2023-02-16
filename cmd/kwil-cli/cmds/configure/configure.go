package configure

import (
	"fmt"
	"kwil/cmd/kwil-cli/cmds/common"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/crypto"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmdConfigure() *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "configure",
		Short:         "Configure your client",
		Long:          "",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.LoadConfig()

			runner := &configPrompter{
				Viper: viper.GetViper(),
			}

			// endpoint
			runner.AddPrompt(&common.Prompter{
				Label:   "Kwil RPC URL",
				Default: viper.GetString(config.KwilProviderRpcUrlKey),
			}, config.KwilProviderRpcUrlKey, removeProtocol)

			// private key
			runner.AddPrompt(&common.Prompter{
				Label:   "Private Key",
				Default: viper.GetString(config.WalletPrivateKeyKey),
			}, config.WalletPrivateKeyKey, isValidPrivateKey)

			// eth provider
			runner.AddPrompt(&common.Prompter{
				Label:   "Ethereum RPC URL",
				Default: viper.GetString(config.ClientChainProviderRpcUrlKey),
			}, config.ClientChainProviderRpcUrlKey, containsProtocol)

			// run the prompts
			if err := runner.Run(); err != nil {
				return err
			}

			return viper.WriteConfig()
		},
	}

	return cmd
}

// removeProtocol should remove the http:// or https:// from the url
func removeProtocol(url *string) error {
	*url = strings.Replace(*url, "http://", "", 1)
	*url = strings.Replace(*url, "https://", "", 1)
	*url = strings.Replace(*url, "ws://", "", 1)
	*url = strings.Replace(*url, "wss://", "", 1)

	return nil
}

// containsProtocol should check if the url contains http:// or https://
func containsProtocol(url *string) error {
	if strings.Contains(*url, "http://") || strings.Contains(*url, "https://") || strings.Contains(*url, "ws://") || strings.Contains(*url, "wss://") {
		return nil
	}
	return fmt.Errorf("url must contain http:// or https://")
}

func isValidPrivateKey(pk *string) error {
	_, err := crypto.ECDSAFromHex(*pk)
	if err != nil {
		return fmt.Errorf(`invalid private key.  key could not be converted to hex: %w`, err)
	}
	return nil
}
