package app

import (
	"fmt"
	"os"
	"path/filepath"

	"kwil/node"

	"github.com/spf13/cobra"
)

func ResetCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the blockchain and the application state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, err := RootDir(cmd)
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

			// remove state.json file if it exists
			stateFile := filepath.Join(rootDir, "state.json")
			if _, err := os.Stat(stateFile); err == nil {
				if err := os.Remove(stateFile); err != nil {
					return err
				}
				fmt.Println("state.json removed")
			}

			// remove the blockstore if all is set
			if all {
				chainDir := filepath.Join(rootDir, "blockstore")
				if _, err := os.Stat(chainDir); err == nil {
					if err := os.RemoveAll(chainDir); err != nil {
						return err
					}
					fmt.Println("blockstore removed")
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "reset all data, if this is not set, only the app state will be reset")

	return cmd
}
