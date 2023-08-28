package utils

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	//cfg "github.com/cometbft/cometbft/config"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/abci"
)

// ResetAllCmd removes the database of this CometBFT core
// instance.

var keepAddrBook bool

func NewResetAllCmd() *cobra.Command {
	// XXX: this is totally unsafe.
	// it's only suitable for testnets.
	cmd := &cobra.Command{
		Use:     "unsafe-reset-all",
		Aliases: []string{"unsafe_reset_all"},
		Short:   "(unsafe) Remove all the blockchain's data and WAL, reset this node's validator to genesis state, for testing purposes only",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			homeDir := viper.GetString("home")
			configFile := filepath.Join(homeDir, "abci", "config", "config.toml")
			cfg := config.DefaultConfig()
			err = cfg.ParseConfig(configFile)
			if err != nil {
				return err
			}

			// TODO: this probably will fail since we no longer have those key files.
			return abci.ResetAll(
				cfg.ChainCfg.DBDir(),
				cfg.ChainCfg.P2P.AddrBookFile(),
				cfg.ChainCfg.PrivValidatorKeyFile(),
				cfg.ChainCfg.PrivValidatorStateFile(),
			)
		},
	}

	cmd.Flags().BoolVar(&keepAddrBook, "keep-addr-book", false, "keep the address book intact")
	return cmd
}

// ResetStateCmd removes the database of the specified CometBFT core instance.
func NewResetStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reset-state",
		Aliases: []string{"reset_state"},
		Short:   "(unsafe) Remove all the data and WAL, for testing purposes only",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			homeDir := viper.GetString("home")
			configFile := filepath.Join(homeDir, "abci", "config", "config.toml")
			cfg := config.DefaultConfig()
			err = cfg.ParseConfig(configFile)
			if err != nil {
				return err
			}
			return abci.ResetState(cfg.ChainCfg.DBDir())
		},
	}

	return cmd
}
