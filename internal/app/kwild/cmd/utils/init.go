package utils

import (
	"os"

	"github.com/kwilteam/kwil-db/internal/pkg/nodecfg"
	"github.com/spf13/cobra"
)

var initFlags nodecfg.NodeGenerateConfig

// InitFilesCmd initializes a fresh CometBFT instance.
func InitFilesCmd() *cobra.Command {
	var initFilesCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes files required for a kwil node",
		RunE:  initFiles,
	}
	initFilesCmd.Flags().StringVar(&initFlags.HomeDir, "home", "", "comet home directory")
	// TODO: let viper handle this
	if initFlags.HomeDir == "" {
		initFlags.HomeDir = os.Getenv("COMET_BFT_HOME")
	}
	initFilesCmd.Flags().Int64Var(&initFlags.InitialHeight, "initial-height", 0, "initial height of the first block")

	return initFilesCmd
}

func initFiles(cmd *cobra.Command, args []string) error {
	return nodecfg.GenerateNodeConfig(&initFlags)
}
