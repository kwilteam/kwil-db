package validators

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/spf13/cobra"
)

var (
	joinStatusLong    = `Query the status of a pending validator join request.`
	joinStatusExample = `# Query the status of a pending validator join request, by hex public key
kwil-admin validators join-status 6ecaca8e9394c939a858c2c7b47acb1db26a96d7ab38bd702fa3820c5034e9d0`
)

func joinStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "join-status <joiner>",
		Short:   "Query the status of a pending validator join request.",
		Long:    joinStatusLong,
		Example: joinStatusExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			clt, err := common.GetAdminSvcClient(ctx, cmd)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			pubkeyBts, err := hex.DecodeString(args[0])
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			data, err := clt.JoinStatus(ctx, pubkeyBts)
			if err != nil {
				if errors.Is(err, client.ErrNotFound) {
					return display.PrintErr(cmd, errors.New("no active join request for that validator"))
				}
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, &respValJoinStatus{Data: data})
		},
	}

	return cmd
}

// respValJoinStatus represent the status of a validator join request in cli
type respValJoinStatus struct {
	Data *types.JoinRequest
}

// respValJoinRequest is customized json format for respValJoinStatus
type respValJoinRequest struct {
	Candidate string   `json:"candidate"`
	Power     int64    `json:"power"`
	Board     []string `json:"board"`
	Approved  []bool   `json:"approved"`
}

func (r *respValJoinStatus) MarshalJSON() ([]byte, error) {
	joinReq := &respValJoinRequest{
		Candidate: fmt.Sprintf("%x", r.Data.Candidate),
		Power:     r.Data.Power,
		Board:     make([]string, len(r.Data.Board)),
		Approved:  r.Data.Approved,
	}
	for i := range r.Data.Board {
		joinReq.Board[i] = fmt.Sprintf("%x", r.Data.Board[i])
	}

	return json.Marshal(joinReq)
}

func (r *respValJoinStatus) MarshalText() ([]byte, error) {
	approved := 0
	for _, a := range r.Data.Approved {
		if a {
			approved++
		}
	}

	needed := int(math.Ceil(float64(len(r.Data.Board)) * 2 / 3))

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("Candidate: %x\n", r.Data.Candidate))
	msg.WriteString(fmt.Sprintf("Requested Power: %d\n", r.Data.Power))
	msg.WriteString(fmt.Sprintf("Expiration Height: %d\n", r.Data.ExpiresAt))

	msg.WriteString(fmt.Sprintf("%d Approvals Received (%d needed):\n", approved, needed))

	for i := range r.Data.Board {
		approvedTerm := "approved"
		if !r.Data.Approved[i] {
			approvedTerm = "not approved"
		}

		msg.WriteString(fmt.Sprintf("  Validator %x, %s\n",
			r.Data.Board[i], approvedTerm))
	}

	return msg.Bytes(), nil
}
