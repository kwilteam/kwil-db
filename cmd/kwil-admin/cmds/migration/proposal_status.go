package migration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/spf13/cobra"
)

var (
	statusExample = `# Get the status of the pending migration.
kwil-admin migrate proposal-status <proposal_id>`
)

func proposalStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "proposal-status",
		Short:   "Get the status of the pending migration proposal.",
		Long:    "Get the status of the pending migration proposal.",
		Example: statusExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			proposalID, err := types.ParseUUID(args[0])
			if err != nil {
				return display.PrintErr(cmd, err)
			}
			status, err := clt.ResolutionStatus(ctx, proposalID)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &MigrationStatus{
				ProposalID: status.ResolutionID,
				ExpiresAt:  status.ExpiresAt,
				Board:      status.Board,
				Approved:   status.Approved,
			})
		},
	}
}

type MigrationStatus struct {
	ProposalID *types.UUID
	ExpiresAt  int64    `json:"expires_at"` // ExpiresAt is the block height at which the migration proposal expires
	Board      [][]byte `json:"board"`      // Board is the list of validators who are eligible to vote on the migration proposal
	Approved   []bool   `json:"approved"`   // Approved is the list of bools indicating if the corresponding validator approved the migration proposal
}

func (m *MigrationStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(m)
}

func (m *MigrationStatus) MarshalText() ([]byte, error) {
	approved := 0
	for _, a := range m.Approved {
		if a {
			approved++
		}
	}
	needed := int(math.Ceil(float64(len(m.Board)) * 2 / 3))

	var msg bytes.Buffer
	msg.WriteString("Migration Status:\n")
	msg.WriteString(fmt.Sprintf("\tProposal ID: %s\n", m.ProposalID.String()))
	msg.WriteString(fmt.Sprintf("\tExpiresAt: %d\n", m.ExpiresAt))
	msg.WriteString(fmt.Sprintf("\tApprovals Received: %d (needed %d)\n", approved, needed))

	for i := range m.Board {
		status := "not approved"
		if m.Approved[i] {
			status = "approved"
		}

		msg.WriteString(fmt.Sprintf("\t\tValidator %x: (%s)\n", m.Board[i], status))
	}

	return msg.Bytes(), nil
}
