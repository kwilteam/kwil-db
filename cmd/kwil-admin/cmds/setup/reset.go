package setup

import (
	"errors"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/spf13/cobra"
)

var (
	resetLong = `To delete all of a Kwil node's data files, use the ` + "`" + `reset` + "`" + ` command. If directories are not specified, the node's default directories will be used.

WARNING: This command should not be used on production systems. This should only be used to reset disposable test nodes.`

	resetExample = `# Delete all of a Kwil node's data files
kwil-admin setup reset --root-dir "~/.kwild" --sqlpath "~/.kwild/data/kwil.db"`
)

func resetCmd() *cobra.Command {
	var rootDir, snapPath string
	var force bool

	cmd := &cobra.Command{
		Use:     "reset",
		Short:   "To delete all of a Kwil node's data files, use the `reset` command.",
		Long:    resetLong,
		Example: resetExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootDir == "" {
				if !force {
					return display.PrintErr(cmd, errors.New("not removing default home directory without --force or --root-dir"))
				}
				rootDir = common.DefaultKwildRoot()
			}

			if snapPath == "" {
				snapPath = config.SnapshotDirName
			}

			expandedRoot, err := expandPath(rootDir)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			err = config.ResetAll(expandedRoot, snapPath)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&rootDir, "root-dir", "r", "", "root directory of the kwild node")
	// cmd.Flags().StringVarP(&sqlPath, "sqlpath", "s", "", "path to the SQLite database") // TODO: postgres config
	cmd.Flags().StringVarP(&snapPath, "snappath", "p", "", "path to the snapshot directory")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force removal of default home directory")

	return cmd
}
