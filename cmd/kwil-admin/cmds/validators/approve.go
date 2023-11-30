package validators

import (
	"context"
	"encoding/hex"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var (
	approveLong = "A current validator may approve an active join request for a candidate validator using the `approve` subcommand. If enough validators approve the join request, the candidate validator will be added to the validator set."

	approveExample = `# Approve a join request for a candidate validator
kwil-admin validators approve 6ecaca8e9394c939a858c2c7b47acb1db26a96d7ab38bd702fa3820c5034e9d0`
)

func approveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "approve <joiner>",
		Short:   "A current validator may approve an active join request for a candidate validator using the `approve` subcommand.",
		Long:    approveLong,
		Example: approveExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return err
			}

			joinerBts, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			txHash, err := clt.Approve(ctx, joinerBts)
			if err != nil {
				return err
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))
		},
	}

	return cmd
}
