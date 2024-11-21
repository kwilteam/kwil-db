package validators

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

const validatorsShort = "The `validators` command provides functions for creating and broadcasting validator-related transactions."
const validatorsLong = "The `validators` command provides functions for creating and broadcasting validator-related transactions (join/approve/leave), and retrieving information on the current validators and join requests."

var validatorsCmd = &cobra.Command{
	Use:   "validators",
	Short: validatorsShort,
	Long:  validatorsLong,
}

func NewValidatorsCmd() *cobra.Command {
	validatorsCmd.AddCommand(
		joinCmd(),
		joinStatusCmd(),
		listCmd(),
		approveCmd(),
		removeCmd(),
		leaveCmd(),
		listJoinRequestsCmd(),
	)

	common.BindRPCFlags(validatorsCmd)

	return validatorsCmd
}
