package main

import (
	"fmt"
	"os"
	"p2p/node"
	"path/filepath"

	"github.com/spf13/cobra"
)

func resetCmd() *cobra.Command {
	var rootDir string
	var all bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the blockchain and the application state",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir, err := node.ExpandPath(rootDir)
			if err != nil {
				return err
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

	cmd.Flags().StringVar(&rootDir, "root", ".testnet", "root directory for the configuration")
	cmd.Flags().BoolVar(&all, "all", false, "reset all data, if this is not set, only the app state will be reset")

	return cmd
}
