package utils

import (
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

	initFilesCmd.Flags().StringVarP(&initFlags.OutputDir, "output-dir", "o", ".kwild",
		"directory to store initialization data for the node")
	initFilesCmd.Flags().Int64VarP(&initFlags.InitialHeight, "initial-height", "i", 0,
		"initial height of the first block")
	return initFilesCmd
}

func initFiles(cmd *cobra.Command, args []string) error {
	return nodecfg.GenerateNodeConfig(&initFlags)
}
