package configure

import (
	"fmt"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmdConfigure() *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "configure",
		Short:         "Configure your client",
		Long:          "",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadPersistedConfig()
			if err != nil {
				return err
			}

			promptGRPCURL(conf)
			promptPrivateKey(conf)
			promptClientChainRPCURL(conf)

			return config.PersistConfig(conf)
		},
	}

	return cmd
}

func promptGRPCURL(conf *config.KwilCliConfig) {
	prompt := &common.Prompter{
		Label:   "Kwil RPC URL",
		Default: conf.GrpcURL,
	}
	res, err := prompt.Run()
	if err != nil {
		panic(err)
	}

	conf.GrpcURL = res
}

func promptPrivateKey(conf *config.KwilCliConfig) {
	prompt := &common.Prompter{
		Label:   "Private Key",
		Default: crypto.HexFromECDSAPrivateKey(conf.PrivateKey),
	}
	res, err := prompt.Run()
	if err != nil {
		panic(err)
	}

	pk, err := crypto.ECDSAFromHex(res)
	if err != nil {
		fmt.Println(`invalid private key.  key could not be converted to hex.  received: `, res)
		promptPrivateKey(conf)
		return
	}

	conf.PrivateKey = pk
}

func promptClientChainRPCURL(conf *config.KwilCliConfig) {
	prompt := &common.Prompter{
		Label:   "Client Chain RPC URL",
		Default: conf.ClientChainRPCURL,
	}
	res, err := prompt.Run()
	if err != nil {
		panic(err)
	}

	if containsProtocol(&res) != nil {
		fmt.Println(`url must contain http:// , https:// , ws:// , or wss://.  received: `, res)
		promptClientChainRPCURL(conf)
		return
	}

	conf.ClientChainRPCURL = res
}

// containsProtocol should check if the url contains http:// or https://
func containsProtocol(url *string) error {
	if strings.Contains(*url, "http://") || strings.Contains(*url, "https://") || strings.Contains(*url, "ws://") || strings.Contains(*url, "wss://") {
		return nil
	}
	return fmt.Errorf("url must contain http:// or https://")
}
