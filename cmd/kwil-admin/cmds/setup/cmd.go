package setup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const setupLong = `The ` + "`" + `setup` + "`" + ` command provides functions for creating and managing node configuration and data, including:
	- performing quick setup of a standalone Kwil node (init) and Kwil test networks (testnet)
	- updating genesis config with initial SQLite files (genesis-hash)
	- resetting node state and all data files (reset)`

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "The `setup` command provides functions for creating and managing node configuration and data.",
	Long:  setupLong,
}

func NewSetupCmd() *cobra.Command {
	setupCmd.AddCommand(
		initCmd(),
		resetCmd(),
		testnetCmd(),
		genesisHashCmd(),
		resetStateCmd(),
	)

	return setupCmd
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}
	return filepath.Abs(path)
}
