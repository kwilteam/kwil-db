package migration

import (
	"context"
	"errors"
	"math/big"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/spf13/cobra"
)

var (
	triggerLong = "A current validator may trigger a migration request using the `trigger` subcommand. This will create a migration resolution which the validators will vote on. If enough validators approve the migration request, the migration will start after activationHeight number of blocks since approval and stay for the migrationDuration number of blocks."

	triggerExample = `# Trigger a migration request which will start after 1000 blocks after approval and last for 14400 blocks (1 day)
kwil-admin migration trigger 1000 14400 kwil-chain-new`
)

func triggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "trigger <activationHeight> <migrationDuration> <chainID>",
		Short:   "Trigger a migration transaction.",
		Long:    triggerLong,
		Example: triggerExample,
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// convert args[0] string to *big.Int
			activationHeight, ok := new(big.Int).SetString(args[0], 10)
			if !ok {
				return display.PrintErr(cmd, errors.New("failed to convert activationHeight to big.Int"))
			}

			// convert args[1] string to *big.Int
			migrationDuration, ok := new(big.Int).SetString(args[1], 10)
			if !ok {
				return display.PrintErr(cmd, errors.New("failed to convert migrationDuration to big.Int"))
			}

			txHash, err := clt.TriggerMigration(ctx, activationHeight, migrationDuration, args[2])
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))

		},
	}

	return cmd
}
