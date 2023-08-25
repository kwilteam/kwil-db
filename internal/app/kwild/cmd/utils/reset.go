package utils

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/kwilteam/kwil-db/pkg/abci"
)

// ResetAllCmd removes the database of this CometBFT core
// instance.

var keepAddrBook bool

func NewResetAllCmd() *cobra.Command {
	var homeDir string

	// XXX: this is totally unsafe.
	// it's only suitable for testnets.
	cmd := &cobra.Command{
		Use:     "unsafe-reset-all",
		Aliases: []string{"unsafe_reset_all"},
		Short:   "(unsafe) Remove all the blockchain's data and WAL, reset this node's validator to genesis state, for testing purposes only",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			config, err := ParseConfig(cmd, homeDir)
			if err != nil {
				return err
			}

			return abci.ResetAll(
				config.DBDir(),
				config.P2P.AddrBookFile(),
				config.PrivValidatorKeyFile(),
				config.PrivValidatorStateFile(),
			)
		},
	}

	cmd.Flags().StringVar(&homeDir, "home", "", "comet home directory")
	// TODO: let viper handle this
	if homeDir == "" {
		homeDir = os.Getenv("COMET_BFT_HOME")
	}
	cmd.Flags().BoolVar(&keepAddrBook, "keep-addr-book", false, "keep the address book intact")
	return cmd
}

// ResetStateCmd removes the database of the specified CometBFT core instance.
func NewResetStateCmd() *cobra.Command {
	var homeDir string

	cmd := &cobra.Command{
		Use:     "reset-state",
		Aliases: []string{"reset_state"},
		Short:   "(unsafe) Remove all the data and WAL, for testing purposes only",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			conf, err := ParseConfig(cmd, homeDir)
			if err != nil {
				return err
			}

			return abci.ResetState(conf.DBDir())
		},
	}

	cmd.Flags().StringVar(&homeDir, "home", "", "comet home directory")
	if homeDir == "" {
		homeDir = os.Getenv("COMET_BFT_HOME")
	}
	return cmd
}

func ParseConfig(cmd *cobra.Command, homeDir string) (*cfg.Config, error) {
	conf := cfg.DefaultConfig()
	conf.SetRoot(homeDir)
	cfg.EnsureRoot(conf.RootDir)

	if err := conf.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}
	if warnings := conf.CheckDeprecated(); len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Println("deprecated usage found in configuration file", "usage", warning)
		}
	}
	return conf, nil
}
