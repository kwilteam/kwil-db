package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/p2p"
)

// GenNodeKeyCmd allows the generation of a node key. It prints node's ID to
// the standard output.
func GenNodeKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gen-node-key",
		Aliases: []string{"gen_node_key"},
		Short:   "Generate a node key for this node and print its ID",
		RunE:    genNodeKey,
	}
	cmd.Flags().StringVar(&outputDir, "home", "", "comet home directory")
	if outputDir == "" {
		outputDir = os.Getenv("COMET_BFT_HOME")
	}
	return cmd
}

func genNodeKey(cmd *cobra.Command, args []string) error {
	nodeKey, err := p2p.LoadOrGenNodeKey(filepath.Join(outputDir, "config/node_key.json"))
	if err != nil {
		return err
	}
	fmt.Println(nodeKey.ID())
	return nil
}
