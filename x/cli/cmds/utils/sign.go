package utils

import (
	"fmt"
	"kwil/x/cli/chain"
	"kwil/x/crypto"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func signCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sign",
		Short: "Sign is used to generate a signature for a given message.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			// check there is 1 arg
			if len(args) != 1 {
				return fmt.Errorf("sign requires one argument: message")
			}

			// get private key
			c, err := chain.NewClientV(viper.GetViper())
			if err != nil {
				return fmt.Errorf("error getting config: %w", err)
			}

			// generate signature
			sig, err := crypto.Sign([]byte(args[0]), c.PrivateKey)
			if err != nil {
				return fmt.Errorf("error generating signature: %w", err)
			}

			fmt.Println(sig)
			return nil
		},
	}

	return cmd
}
