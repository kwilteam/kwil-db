package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/app/shared/bind"
	"github.com/kwilteam/kwil-db/node"

	"github.com/spf13/cobra"
)

func ResetCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the blockchain and the application state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, err := bind.RootDir(cmd)
			if err != nil {
				return err // the parent command needs to set a persistent flag named "root"
			}
			rootDir, err = node.ExpandPath(rootDir)
			if err != nil {
				return err
			}
			if _, err := os.Stat(rootDir); os.IsNotExist(err) {
				return fmt.Errorf("root directory %s does not exist", rootDir)
			}

			// TODO: reset app DB

			if !all {
				return nil
			}

			// remove the blockstore if all is set
			chainDir := filepath.Join(rootDir, "blockstore")
			if err := os.RemoveAll(chainDir); err != nil {
				return err
			}
			fmt.Println("blockstore removed")

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "reset all data, if this is not set, only the app state will be reset")

	return cmd
}
