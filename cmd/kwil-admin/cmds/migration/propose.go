package migration

import (
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/internal/migrations"
	"github.com/kwilteam/kwil-db/internal/voting"
)

var (
	proposeLong = `A validator operator can submit a migration proposal using the ` + "`" + `propose` + "`" + ` subcommand.
	
The migration proposal includes the ` + "`" + `activation-period` + "`" + ` and ` + "`" + `duration` + "`" + `. This will generate a migration resolution
for the other validators to vote on. If a super-majority of validators approve the migration proposal, the migration will
commence after the specified activation-period blocks from approval and will continue for the duration defined by duration blocks.`

	proposeExample = `# Submit a migration proposal to migrate to a new chain with activation period 1000 and migration duration of 14400 blocks.
kwil-admin migrate propose --activation-period 1000 --duration 14400
(or)
kwil-admin migrate propose -a 1000 -d 14400`
)

func proposeCmd() *cobra.Command {
	var activationPeriod, migrationDuration uint64

	cmd := &cobra.Command{
		Use:     "propose",
		Short:   "Submit a migration proposal.",
		Long:    proposeLong,
		Example: proposeExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if migrationDuration <= 0 || activationPeriod <= 0 {
				return display.PrintErr(cmd, errors.New("start-height and migration duration must be greater than 0"))
			}

			proposal := migrations.MigrationDeclaration{
				ActivationPeriod: activationPeriod,
				Duration:         migrationDuration,
				Timestamp:        time.Now().String(),
			}
			proposalBts, err := proposal.MarshalBinary()
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			// Submit a migration proposal
			txHash, err := clt.CreateResolution(ctx, proposalBts, voting.StartMigrationEventType)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespTxHash(txHash))

		},
	}

	cmd.Flags().Uint64VarP(&activationPeriod, "activation-period", "a", 0, "The number of blocks before the migration is activated since the approval of the proposal.")
	cmd.Flags().Uint64VarP(&migrationDuration, "duration", "d", 0, "The duration of the migration.")
	return cmd
}
