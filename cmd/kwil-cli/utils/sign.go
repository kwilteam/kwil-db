package utils

import (
	"fmt"
	"kwil/x/crypto"
	"kwil/x/fund"

	"github.com/spf13/cobra"
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
			config, err := fund.NewConfig()
			if err != nil {
				return fmt.Errorf("error getting client config: %w", err)
			}

			// generate signature
			sig, err := crypto.Sign([]byte(args[0]), config.PrivateKey)
			if err != nil {
				return fmt.Errorf("error generating signature: %w", err)
			}

			fmt.Println(sig)
			return nil
		},
	}

	return cmd
}
