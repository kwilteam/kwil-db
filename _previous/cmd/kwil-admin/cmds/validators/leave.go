package validators

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var (
	leaveLong = "A current validator may leave the validator set using the `leave` command."

	leaveExample = `# Leave the validator set
kwil-admin validators leave`
)

func leaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "leave",
		Short:   "A current validator may leave the validator set using the `leave` command.",
		Long:    leaveLong,
		Example: leaveExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			txHash, err := clt.Leave(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))
		},
	}

	return cmd
}
