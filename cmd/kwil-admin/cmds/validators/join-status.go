package validators

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/cmds/common"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/internal/validators"
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
				return err
			}

			pubkeyBts, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			data, err := clt.JoinStatus(ctx, pubkeyBts)
			if err != nil {
				if errors.Is(err, client.ErrNotFound) {
					return errors.New("no active join request for that validator")
				}
				return err
			}

			return display.PrintCmd(cmd, &respValJoinStatus{Data: data})
		},
	}

	return cmd
}

// respValJoinStatus represent the status of a validator join request in cli
type respValJoinStatus struct {
	Data *validators.JoinRequest
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
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("Candidate: %x (want power %d)\n",
		r.Data.Candidate, r.Data.Power))
	for i := range r.Data.Board {
		msg.WriteString(fmt.Sprintf(" Validator %x, approved = %v\n",
			r.Data.Board[i], r.Data.Approved[i]))
	}

	return msg.Bytes(), nil
}
