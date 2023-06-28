package utils

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
)

// InitFilesCmd initializes a fresh CometBFT instance.
func InitFilesCmd() *cobra.Command {
	var initFilesCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes files required for a kwil node",
		RunE:  initFiles,
	}
	initFilesCmd.Flags().StringVar(&outputDir, "home", "", "comet home directory")
	if outputDir == "" {
		outputDir = os.Getenv("COMET_BFT_HOME")
	}
	initFilesCmd.Flags().BoolVar(&disable_gas, "disable-gas", false,
		"Disables gas costs on all transactions and once the network is initialized, it can't be changed")
	return initFilesCmd
}
func initFiles(cmd *cobra.Command, args []string) error {
	config := cfg.DefaultConfig()
	config.SetRoot(outputDir)
	err := os.MkdirAll(filepath.Join(outputDir, "config"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(outputDir)
		return err
	}

	err = os.MkdirAll(filepath.Join(outputDir, "data"), nodeDirPerm)
	if err != nil {
		_ = os.RemoveAll(outputDir)
		return err
	}
	return InitFilesWithConfig(config)
}
