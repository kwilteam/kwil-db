package setup

import (
	"errors"

	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/spf13/cobra"
)

var (
	resetStateLong = `Delete blockchain state files.`

	resetStateExample = `# Delete blockchain state files
kwil-admin setup reset-state --root_dir "~/.kwild"`
)

func resetStateCmd() *cobra.Command {
	var rootDir string
	var force bool

	cmd := &cobra.Command{
		Use:     "reset-state",
		Short:   "Delete blockchain state files.",
		Long:    resetStateLong,
		Example: resetStateExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootDir == "" {
				if !force {
					return errors.New("not removing default home directory without --force or --root_dir")
				}
				rootDir = common.DefaultKwildRoot()
			}

			expandedDir, err := expandPath(rootDir)
			if err != nil {
				return err
			}

			return config.ResetChainState(expandedDir)
		},
	}

	cmd.Flags().StringVarP(&rootDir, "root_dir", "r", "", "root directory of the kwild node")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force removal of default home directory")

	return cmd
}
