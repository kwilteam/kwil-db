package utils

import (
	"encoding/hex"

	"github.com/cometbft/cometbft/crypto/ed25519"

	"github.com/spf13/cobra"
)

func GenPrivateKeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gen-private-key",
		Short: "Generates ed25519 private key and shows the pubkey, CometBFT address, and node ID for the generated private key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			privateKey := ed25519.GenPrivKey()
			showKeyInfo(hex.EncodeToString(privateKey[:]))
			return nil
		},
	}
}
