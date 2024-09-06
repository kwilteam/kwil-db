package setup

import (
	"errors"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/spf13/cobra"
)

var (
	resetLong = `To delete all of a Kwil node's data files, use the ` + "`" + `reset` + "`" + ` command. If directories are not specified, the node's default directories will be used.

WARNING: This command should not be used on production systems. This should only be used to reset disposable test nodes.`

	resetExample = `# Delete all of a Kwil node's data files
kwil-admin setup reset --root-dir "~/.kwild"`
)

func resetCmd() *cobra.Command {
	var rootDir, snapPath string
	var force bool

	resetCmd := &cobra.Command{
		Use:     "reset",
		Short:   "To delete all of a Kwil node's data files, use the `reset` command.",
		Long:    resetLong,
		Example: resetExample,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if rootDir == "" {
				if !force {
					return display.PrintErr(cobraCmd, errors.New("not removing default home directory without --force or --root-dir"))
				}
				rootDir = common.DefaultKwildRoot()
			}

			if snapPath == "" {
				snapPath = cmd.DefaultConfig().AppConfig.Snapshots.SnapshotDir
			}

			expandedRoot, err := common.ExpandPath(rootDir)
			if err != nil {
				return display.PrintErr(cobraCmd, err)
			}

			err = config.ResetAll(expandedRoot, snapPath)
			if err != nil {
				return display.PrintErr(cobraCmd, err)
			}

			return nil
		},
	}

	resetCmd.Flags().StringVarP(&rootDir, "root-dir", "r", "", "root directory of the kwild node")
	resetCmd.Flags().StringVarP(&snapPath, "snappath", "p", "", "path to the snapshot directory")
	resetCmd.Flags().BoolVarP(&force, "force", "f", false, "force removal of default home directory")

	return resetCmd
}
