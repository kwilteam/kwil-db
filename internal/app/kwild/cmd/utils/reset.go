package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/privval"
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

		return resetState(conf.DBDir())
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

	return resetAll(
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

	resetFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
	return nil
}

// resetAll removes address book files plus all data, and resets the privValdiator data.
func resetAll(dbDir, addrBookFile, privValKeyFile, privValStateFile string) error {
	if keepAddrBook {
		fmt.Println("The address book remains intact")
	} else {
		removeAddrBook(addrBookFile)
	}

	if err := os.RemoveAll(dbDir); err == nil {
		fmt.Println("Removed all blockchain history", "dir", dbDir)
	} else {
		fmt.Println("Error removing all blockchain history", "dir", dbDir, "err", err)
	}

	if err := cmtos.EnsureDir(dbDir, 0700); err != nil {
		fmt.Println("Error recreating dbDir", "dir", dbDir, "err", err)
	}

	// recreate the dbDir since the privVal state needs to live there
	resetFilePV(privValKeyFile, privValStateFile)
	return nil
}

// resetState removes address book files plus all databases.
func resetState(dbDir string) error {
	blockdb := filepath.Join(dbDir, "blockstore.db")
	state := filepath.Join(dbDir, "state.db")
	wal := filepath.Join(dbDir, "cs.wal")
	evidence := filepath.Join(dbDir, "evidence.db")
	txIndex := filepath.Join(dbDir, "tx_index.db")

	if cmtos.FileExists(blockdb) {
		if err := os.RemoveAll(blockdb); err == nil {
			fmt.Println("Removed all blockstore.db", "dir", blockdb)
		} else {
			fmt.Println("error removing all blockstore.db", "dir", blockdb, "err", err)
		}
	}

	if cmtos.FileExists(state) {
		if err := os.RemoveAll(state); err == nil {
			fmt.Println("Removed all state.db", "dir", state)
		} else {
			fmt.Println("error removing all state.db", "dir", state, "err", err)
		}
	}

	if cmtos.FileExists(wal) {
		if err := os.RemoveAll(wal); err == nil {
			fmt.Println("Removed all cs.wal", "dir", wal)
		} else {
			fmt.Println("error removing all cs.wal", "dir", wal, "err", err)
		}
	}

	if cmtos.FileExists(evidence) {
		if err := os.RemoveAll(evidence); err == nil {
			fmt.Println("Removed all evidence.db", "dir", evidence)
		} else {
			fmt.Println("error removing all evidence.db", "dir", evidence, "err", err)
		}
	}

	if cmtos.FileExists(txIndex) {
		if err := os.RemoveAll(txIndex); err == nil {
			fmt.Println("Removed all tx_index.db", "dir", txIndex)
		} else {
			fmt.Println("error removing all tx_index.db", "dir", txIndex, "err", err)
		}
	}

	if err := cmtos.EnsureDir(dbDir, 0700); err != nil {
		fmt.Println("unable to recreate dbDir", "err", err)
	}
	return nil
}

func resetFilePV(privValKeyFile, privValStateFile string) {
	if _, err := os.Stat(privValKeyFile); err == nil {
		pv := privval.LoadFilePVEmptyState(privValKeyFile, privValStateFile)
		pv.Reset()
		fmt.Println("Reset private validator file to genesis state", "keyFile", privValKeyFile, "stateFile", privValStateFile)
	} else {
		pv := privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
		fmt.Println("Generated private validator file", "keyFile", privValKeyFile, "stateFile", privValStateFile)
	}
}

func removeAddrBook(addrBookFile string) {
	if err := os.Remove(addrBookFile); err == nil {
		fmt.Println("Removed existing address book", "file", addrBookFile)
	} else if !os.IsNotExist(err) {
		fmt.Println("Error removing address book", "file", addrBookFile, "err", err)
	}
}
