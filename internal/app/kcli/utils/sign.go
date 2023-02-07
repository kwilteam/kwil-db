package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"kwil/internal/app/kcli/config"
	"kwil/pkg/crypto"
)

func signCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sign",
		Short: "Sign is used to generate a signature for a given message.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// generate signature
			sig, err := crypto.Sign([]byte(args[0]), config.AppConfig.Fund.Wallet)
			if err != nil {
				return fmt.Errorf("error generating signature: %w", err)
			}

			fmt.Println(sig)
			return nil
		},
	}

	return cmd
}
