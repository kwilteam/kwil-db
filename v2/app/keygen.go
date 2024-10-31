package app

import (
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"
)

func KeyCmd() *cobra.Command {
	keyCmd := &cobra.Command{
		Use:   "key",
		Short: "Key management command for testing purposes",
		Run: func(cmd *cobra.Command, args []string) {
			// Logic to generate keys
		},
	}
	keyCmd.AddCommand(KeygenCmd(), KeyInfoCmd())
	return keyCmd
}

func KeygenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gen",
		Short: "Generate secp256k1 key pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Logic to generate key pair
			pk, err := secp256k1.GeneratePrivateKey()
			if err != nil {
				return err
			}

			privKey := (*crypto.Secp256k1PrivateKey)(pk)
			pkRaw, err := privKey.Raw()
			if err != nil {
				return err
			}

			println("Private key(Hex):", hex.EncodeToString(pkRaw))
			return keyInfo(hex.EncodeToString(pkRaw))
		},
	}
}

func KeyInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [private_key]", // TODO: later read from file
		Short: "Get the key from the private key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Logic to get key from private
			return keyInfo(args[0])
		},
	}
}

func keyInfo(privateKey string) error {
	keyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return err
	}

	privKey, err := crypto.UnmarshalSecp256k1PrivateKey(keyBytes)
	if err != nil {
		return err
	}

	pubKey, err := privKey.GetPublic().Raw()
	if err != nil {
		return err
	}

	println("Public key(Hex):", hex.EncodeToString(pubKey))

	nodeID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return err
	}

	println("Node ID:", nodeID.String())
	return nil
}
