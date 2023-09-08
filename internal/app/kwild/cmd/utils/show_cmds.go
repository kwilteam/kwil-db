package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
)

func KeyInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "key-info privateKeyHex",
		Aliases: []string{"key_info"},
		Args:    cobra.ExactArgs(1),
		Short:   "Show the pubkey, CometBFT address, and Node ID for an Ed25519 private key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			showKeyInfo(decodeHexString(args[0]))
			return nil
		},
	}
}

func showKeyInfo(privateKey []byte) {
	priv := ed25519.PrivKey(privateKey)
	pub := priv.PubKey().(ed25519.PubKey)
	nodeID := p2p.PubKeyToID(pub)

	fmt.Printf("Private key (hex): %x\n", priv.Bytes())                                       // KWILD_PRIVATE_KEY ?
	fmt.Printf("Private key (base64): %s\n", base64.StdEncoding.EncodeToString(priv.Bytes())) // "value" in abci/config/node_key.json ?
	fmt.Printf("Public key (base64): %s\n", base64.StdEncoding.EncodeToString(pub.Bytes()))   // "validators.pub_key.value" in abci/config/genesis.json ?
	fmt.Printf("Public key (cometized hex): %v\n", pub.String())                              // for reference with come cometbft logs
	fmt.Printf("Address (string): %s\n", pub.Address().String())                              // "validators.address" in abci/config/genesis.json ?
	fmt.Printf("Node ID: %v\n", nodeID)
	fmt.Printf("Public key (hex): %s\n", hex.EncodeToString(pub.Bytes()))
}

func decodeHexString(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("not hex")
	}
	return b
}
