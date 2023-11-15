package validators

import (
	"context"
	"encoding/hex"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var (
	removeLong = "Command `remove` votes to remove a validator from the validator set. If enough validators vote to remove the validator, the validator will be removed from the validator set."

	removeExample = `$ kwil-admin validators remove e16141e4def3a7f2dfc5bbf40d50619b4d7bc9c9f670fcad98327b0d3d7b97b6 --key-file "~/.kwild/private_key"
TxHash: a001d7ccf73b05d6aa0d749d25cec60fef4de606533ac21f48043c7cc65dfe1b`
)

func removeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <validator>",
		Short:   "Command `remove` votes to remove a validator from the validator set.",
		Long:    removeLong,
		Example: removeExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return err
			}

			validatorBts, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			txHash, err := clt.Remove(ctx, validatorBts)
			if err != nil {
				return err
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))
		},
	}

	return cmd
}
