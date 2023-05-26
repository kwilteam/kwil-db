package configure

import (
	"fmt"
	"strings"

	common "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/prompt"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/crypto"

	"github.com/spf13/cobra"
)

func NewCmdConfigure() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "configure",
		Short: "Configure your client",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadPersistedConfig()
			if err != nil {
				return err
			}

			err = runErrs(conf,
				promptGRPCURL,
				promptPrivateKey,
				promptClientChainRPCURL,
			)
			if err != nil {
				return err
			}

			return config.PersistConfig(conf)
		},
	}

	return cmd
}

func runErrs(conf *config.KwilCliConfig, fns ...func(*config.KwilCliConfig) error) error {
	for _, fn := range fns {

		err := fn(conf)
		if err != nil {
			return err
		}
	}

	return nil
}

func promptGRPCURL(conf *config.KwilCliConfig) error {
	prompt := &common.Prompter{
		Label:   "Kwil RPC URL",
		Default: conf.GrpcURL,
	}
	res, err := prompt.Run()
	if err != nil {
		return err
	}

	conf.GrpcURL = res

	return nil
}

func promptPrivateKey(conf *config.KwilCliConfig) error {
	prompt := &common.Prompter{
		Label:   "Private Key",
		Default: crypto.HexFromECDSAPrivateKey(conf.PrivateKey),
	}
	res, err := prompt.Run()
	if err != nil {
		return err
	}

	pk, err := crypto.ECDSAFromHex(res)
	if err != nil {
		fmt.Println(`invalid private key.  key could not be converted to hex.  received: `, res)
		promptAskAgain := &common.Prompter{
			Label: "Would you like to enter another? (y/n)",
		}
		res2, err := promptAskAgain.Run()
		if err != nil {
			return err
		}

		if res2 == "y" || res == "yes" {
			return promptPrivateKey(conf)
		}

		return nil
	}

	conf.PrivateKey = pk

	return nil
}

func promptClientChainRPCURL(conf *config.KwilCliConfig) error {
	prompt := &common.Prompter{
		Label:   "Client Chain RPC URL",
		Default: conf.ClientChainRPCURL,
	}
	res, err := prompt.Run()
	if err != nil {
		return err
	}

	if containsProtocol(&res) != nil {
		fmt.Println(`url must contain http:// , https:// , ws:// , or wss://.  received: `, res)
		return promptClientChainRPCURL(conf)
	}

	conf.ClientChainRPCURL = res

	return nil
}

// containsProtocol should check if the url contains http:// or https://
func containsProtocol(url *string) error {
	if strings.Contains(*url, "http://") || strings.Contains(*url, "https://") || strings.Contains(*url, "ws://") || strings.Contains(*url, "wss://") {
		return nil
	}
	return fmt.Errorf("url must contain http:// or https://")
}
