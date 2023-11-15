package validators

import (
	"context"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var (
	leaveLong = "A current validator may leave the validator set using the `leave` command."

	leaveExample = `$ kwil-admin validators leave --key-file "~/.kwild/private_key"
TxHash: a001d7ccf73b05d6aa0d749d25cec60fef4de606533ac21f48043c7cc65dfe1b`
)

func leaveCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "leave",
		Short:   "A current validator may leave the validator set using the `leave` command.",
		Long:    leaveLong,
		Example: leaveExample,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return err
			}

			txHash, err := clt.Leave(ctx)
			if err != nil {
				return err
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))
		},
	}

	return cmd
}
