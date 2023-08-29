package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
)

func KeyInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "key-info privateKeyHex",
		Aliases: []string{"key_info"},
		Args:    cobra.ExactArgs(1),
		Short:   "Show the pubkey, CometBFT address, and node ID for an Ed25519 private key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			showKeyInfo(args[0])
			return nil
		},
	}
}

func showKeyInfo(privateKey string) {
	priv := ed25519.PrivKey(decodeHexString(privateKey))
	pub := priv.PubKey().(ed25519.PubKey)
	nodeID := p2p.PubKeyToID(pub)

	fmt.Printf("Private key (hex): %x\n", priv.Bytes())                                       // KWILD_PRIVATE_KEY ?
	fmt.Printf("Private key (base64): %s\n", base64.StdEncoding.EncodeToString(priv.Bytes())) // "value" in abci/config/node_key.json ?
	fmt.Printf("Public key (base64): %s\n", base64.StdEncoding.EncodeToString(pub.Bytes()))   // "validators.pub_key.value" in abci/config/genesis.json ?
	fmt.Printf("Public key (cometized hex): %v\n", pub.String())                              // for reference with come cometbft logs
	fmt.Printf("Address (string): %s\n", pub.Address().String())                              // "validators.address" in abci/config/genesis.json ?
	fmt.Printf("Node ID: %v\n", nodeID)
}

// ShowNodeIDCmd dumps node's ID to the standard output.
func ShowNodeIDCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:     "show-node-id",
		Aliases: []string{"show_node_id"},
		Short:   "Show this node's ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir := viper.GetString("home")
			configFile := filepath.Join(homeDir, "abci", "config", "config.toml")
			cfg := config.DefaultConfig()
			err := cfg.ParseConfig(configFile)
			if err != nil {
				return err
			}

			if cfg.AppCfg.PrivateKey == "" {
				return fmt.Errorf("private key is not set")
			}

			priv := ed25519.PrivKey(decodeHexString(cfg.AppCfg.PrivateKey))
			nodeKey := p2p.NodeKey{PrivKey: priv}
			nodeID := nodeKey.ID()
			fmt.Println("NodeID: ", nodeID)
			return nil
		},
	}

	return &cmd
}

func decodeHexString(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("not hex")
	}
	return b
}
