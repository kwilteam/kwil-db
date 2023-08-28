package utils

import (
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"

	"github.com/spf13/cobra"
)

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
