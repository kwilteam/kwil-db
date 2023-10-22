package configure

import (
	"fmt"

	common "github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/prompt"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto"

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
				promptChainID,
				promptPrivateKey,
				promptTLSCertFile,
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

func promptChainID(conf *config.KwilCliConfig) error {
	prompt := &common.Prompter{
		Label:   "Kwil Chain ID (leave empty to trust a server-provided value)",
		Default: conf.ChainID,
	}
	res, err := prompt.Run()
	if err != nil { // NOTE: empty is valid (no error)
		return err
	}

	conf.ChainID = res

	return nil
}

func promptTLSCertFile(conf *config.KwilCliConfig) error {
	prompt := &common.Prompter{
		Label:   "Kwil RPC TLS certificate path",
		Default: conf.TLSCertFile,
	}
	res, err := prompt.Run()
	if err != nil {
		return err
	}

	conf.TLSCertFile = res

	return nil
}

func promptPrivateKey(conf *config.KwilCliConfig) error {
	var defaultPrivKeyHex string
	if conf.PrivateKey != nil {
		defaultPrivKeyHex = conf.PrivateKey.Hex()
	}
	prompt := &common.Prompter{
		Label:   "Private Key",
		Default: defaultPrivKeyHex,
	}
	res, err := prompt.Run()
	if err != nil {
		return err
	}

	if res == "" {
		conf.PrivateKey = nil
		return nil
	}

	pk, err := crypto.Secp256k1PrivateKeyFromHex(res)
	if err != nil {
		fmt.Printf("invalid private key: %v\n", err)
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
