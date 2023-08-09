package utils

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/kwilteam/kwil-db/pkg/utils"
)

// ResetAllCmd removes the database of this CometBFT core
// instance.

var keepAddrBook bool

func NewResetAllCmd() *cobra.Command {
	ResetAllCmd.Flags().StringVar(&outputDir, "home", "", "comet home directory")
	if outputDir == "" {
		outputDir = os.Getenv("COMET_BFT_HOME")
	}
	ResetAllCmd.Flags().BoolVar(&keepAddrBook, "keep-addr-book", false, "keep the address book intact")
	return ResetAllCmd
}

var ResetAllCmd = &cobra.Command{
	Use:     "unsafe-reset-all",
	Aliases: []string{"unsafe_reset_all"},
	Short:   "(unsafe) Remove all the blockchain's data and WAL, reset this node's validator to genesis state, for testing purposes only",
	RunE:    resetAllCmd,
}

// ResetStateCmd removes the database of the specified CometBFT core instance.
func NewResetStateCmd() *cobra.Command {
	ResetStateCmd.Flags().StringVar(&outputDir, "home", "", "comet home directory")
	if outputDir == "" {
		outputDir = os.Getenv("COMET_BFT_HOME")
	}
	return ResetStateCmd
}

var ResetStateCmd = &cobra.Command{
	Use:     "reset-state",
	Aliases: []string{"reset_state"},
	Short:   "(unsafe) Remove all the data and WAL, for testing purposes only",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		conf, err := ParseConfig(cmd)
		if err != nil {
			return err
		}

		return utils.ResetState(conf.DBDir())
	},
}

// ResetPrivValidatorCmd resets the private validator files.
func NewResetPrivValidatorCmd() *cobra.Command {
	ResetPrivValidatorCmd.Flags().StringVar(&outputDir, "home", "", "comet home directory")
	if outputDir == "" {
		outputDir = os.Getenv("COMET_BFT_HOME")
	}
	return ResetPrivValidatorCmd
}

var ResetPrivValidatorCmd = &cobra.Command{
	Use:     "unsafe-reset-priv-validator",
	Aliases: []string{"unsafe_reset_priv_validator"},
	Short:   "(unsafe) Reset this node's validator to genesis state, for testing purposes only",
	RunE:    resetPrivValidator,
}

// XXX: this is totally unsafe.
// it's only suitable for testnets.
func resetAllCmd(cmd *cobra.Command, args []string) (err error) {
	config, err := ParseConfig(cmd)
	if err != nil {
		return err
	}

	return utils.ResetAll(
		config.DBDir(),
		config.P2P.AddrBookFile(),
		config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(),
	)
}

func ParseConfig(cmd *cobra.Command) (*cfg.Config, error) {
	conf := cfg.DefaultConfig()
	conf.SetRoot(outputDir)
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

// XXX: this is totally unsafe.
// it's only suitable for testnets.
func resetPrivValidator(cmd *cobra.Command, args []string) (err error) {
	config, err := ParseConfig(cmd)
	if err != nil {
		return err
	}

	utils.ResetFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
	return nil
}
