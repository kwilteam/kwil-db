package snapshot

import "github.com/spf13/cobra"

const (
	snapshotExplain = "The `snapshot` command is used to create and manage snapshots."
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: snapshotExplain,
	Long:  snapshotExplain,
}

func NewSnapshotCmd() *cobra.Command {
	snapshotCmd.AddCommand(
		createCmd(),
	)

	return snapshotCmd
}
