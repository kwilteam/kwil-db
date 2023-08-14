package utils

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"

	"github.com/spf13/cobra"
)

func signCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sign",
		Short: "Sign is used to generate a signature for a given message.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadCliConfig()
			if err != nil {
				return err
			}

			if conf.PrivateKey == nil {
				return fmt.Errorf("no private key provided")
			}

			// generate signature
			sig, err := conf.PrivateKey.Sign([]byte(args[0]))
			if err != nil {
				return fmt.Errorf("error generating signature: %w", err)
			}

			fmt.Println(sig)
			return nil
		},
	}

	return cmd
}
