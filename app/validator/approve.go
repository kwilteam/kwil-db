package validator

import (
	"context"
	"encoding/hex"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/shared/display"
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
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := rpc.AdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			joinerBts, err := hex.DecodeString(args[0])
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			txHash, err := clt.Approve(ctx, joinerBts)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))
		},
	}

	return cmd
}
