package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/p2p"

	"github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/privval"
	"github.com/spf13/cobra"
)

// ShowNodeIDCmd dumps node's ID to the standard output.
func ShowNodeIDCmd() *cobra.Command {
	var homeDir string

	cmd := cobra.Command{
		Use:     "show-node-id",
		Aliases: []string{"show_node_id"},
		Short:   "Show this node's ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeKey, err := p2p.LoadNodeKey(filepath.Join(homeDir, "config/node_key.json"))
			if err != nil {
				return err
			}

			fmt.Println(nodeKey.ID())
			return nil
		},
	}

	cmd.Flags().StringVar(&homeDir, "home", "", "comet home directory")
	// TODO: let viper handle this
	if homeDir == "" {
		homeDir = os.Getenv("COMET_BFT_HOME")
	}
	return &cmd
}

func ShowValidatorCmd() *cobra.Command {
	var homeDir string

	cmd := cobra.Command{
		Use:     "show-validator",
		Aliases: []string{"show_validator"},
		Short:   "Show this node's validator info",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyFilePath := filepath.Join(homeDir, "config/priv_validator_key.json")
			if !cmtos.FileExists(keyFilePath) {
				return fmt.Errorf("private validator file %s does not exist", keyFilePath)
			}

			stateFilePath := filepath.Join(homeDir, "data/priv_validator_state.json")
			pv := privval.LoadFilePV(keyFilePath, stateFilePath)
			fmt.Println("Validator pv", pv.Key)

			pubKey, err := pv.GetPubKey()
			if err != nil {
				return fmt.Errorf("can't get pubkey: %w", err)
			}

			fmt.Println("Validator pubkey:", pubKey)

			fmt.Println("Validator address:", pubKey.Address())

			fmt.Println("Validator pubkey address string:", pubKey.Type())
			bz, err := cmtjson.Marshal(pubKey)
			if err != nil {
				return fmt.Errorf("failed to marshal private validator pubkey: %w", err)
			}
			fmt.Println("bz: ", bz)
			fmt.Println(string(bz))

			var publicKey crypto.PubKey
			err = cmtjson.Unmarshal(bz, &publicKey)
			if err != nil {
				return fmt.Errorf("failed to unmarshal private validator pubkey: %w", err)
			}
			fmt.Println("publicKey: ", publicKey)
			return nil
		},
	}

	cmd.Flags().StringVar(&homeDir, "home", "", "comet home directory")
	if homeDir == "" {
		homeDir = os.Getenv("COMET_BFT_HOME")
	}
	return &cmd
}
