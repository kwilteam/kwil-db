package utils

import (
	"fmt"
	"kwil/cmd/kwil-cli/config"
	"kwil/pkg/crypto"

	"github.com/spf13/cobra"
)

func signCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sign",
		Short: "Sign is used to generate a signature for a given message.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pk, err := config.GetEcdsaPrivateKey()
			if err != nil {
				return fmt.Errorf("error getting private key: %w", err)
			}

			// generate signature
			sig, err := crypto.Sign([]byte(args[0]), pk)
			if err != nil {
				return fmt.Errorf("error generating signature: %w", err)
			}

			fmt.Println(sig)
			return nil
		},
	}

	return cmd
}
