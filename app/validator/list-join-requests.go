package validator

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/rpc"
	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/core/types"
)

var (
	listJoinRequestsLong = `Command ` + "`" + `list-join-requests` + "`" + ` lists all pending join requests.
	
Join requests are created when a validator wants to join the validator set. The validator must be approved by 2/3 of the current validator set to be added to the validator set.
Each join request has an expiration block height, after which it is no longer valid.`

	listJoinRequestsExample = `# List all pending join requests
kwil-admin validators list-join-requests`
)

func listJoinRequestsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-join-requests",
		Short:   "List all pending join requests.",
		Long:    listJoinRequestsLong,
		Example: listJoinRequestsExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := rpc.AdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			pending, err := clt.ListPendingJoins(ctx)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &respJoinList{Joins: pending})
		},
	}
	return cmd
}

type respJoinList struct {
	Joins []*types.JoinRequest
}

func (r *respJoinList) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Joins)
}

func (r *respJoinList) MarshalText() ([]byte, error) {
	var msg bytes.Buffer

	if len(r.Joins) == 0 {
		msg.WriteString("No pending join requests")
		return msg.Bytes(), nil
	}

	needed := int(math.Ceil(float64(len(r.Joins[0].Board)) * 2 / 3))

	approvalTerm := "approvals"
	if needed == 1 {
		approvalTerm = "approval"
	}

	// could be ideal to use the SQL table formatting here
	msg.WriteString(fmt.Sprintf("Pending join requests (%d %s needed):\n", needed, approvalTerm))
	msg.WriteString(" Candidate                                                        | Power | Approvals | Expiration\n")
	msg.WriteString("------------------------------------------------------------------+-------+-----------+------------")
	//ref spacing:    22cbbb666c26b2c1f42502df72c32de4d521138a1a2c96121d417a2f341a759c | 1     | 100	   | 100
	for _, j := range r.Joins {
		approvals := 0
		for _, a := range j.Approved {
			if a {
				approvals++
			}
		}

		msg.WriteString(fmt.Sprintf("\n %s | % 5d | % 9d | %d", hex.EncodeToString(j.Candidate), j.Power, approvals, j.ExpiresAt))

	}

	return msg.Bytes(), nil
}
