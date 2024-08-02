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
	proposeLong = "Any validator can submit a migration proposal using the `propose` subcommand. The migration proposal includes the new `chainid`, `activation height` and `migration duration`. This action will generate a migration resolution for the other validators to vote on. If a supermajority of validators approve the migration proposal, the migration will commence after the specified activationHeight blocks from approval and will continue for the duration defined by migrationDuration blocks."

	proposeExample = `# Submit a migration proposal to migrate to a new chain "kwil-chain-new" with activation height 1000 and migration duration of 14400 blocks.
kwil-admin migration propose 1000 14400 kwil-chain-new`
)

func proposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "propose <activationHeight> <migrationDuration> <chainID>",
		Short:   "Submit a migration proposal.",
		Long:    proposeLong,
		Example: proposeExample,
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

			// Submit a migration proposal
			txHash, err := clt.SubmitMigrationProposal(ctx, activationHeight, migrationDuration, args[2])
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))

		},
	}

	return cmd
}
