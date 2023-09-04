package client

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	vmgr "github.com/kwilteam/kwil-db/pkg/validators"
)

func (c *Client) ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*vmgr.JoinRequest, error) {
	resp, err := c.txClient.ValidatorJoinStatus(ctx, &txpb.ValidatorJoinStatusRequest{Pubkey: pubKey})
	if err != nil {
		return nil, fmt.Errorf("failed check validator status: %w", err)
	}
	return convertJoinRequest(pubKey, resp), nil
}

func convertJoinRequest(joiner []byte, resp *txpb.ValidatorJoinStatusResponse) *vmgr.JoinRequest {
	total := len(resp.ApprovedValidators) + len(resp.PendingValidators)
	join := &vmgr.JoinRequest{
		Candidate: joiner,
		Power:     resp.Power,
		Board:     make([][]byte, 0, total),
		Approved:  make([]bool, 0, total),
	}
	for _, vi := range resp.ApprovedValidators {
		join.Board = append(join.Board, vi)
		join.Approved = append(join.Approved, true) // approved
	}
	for _, vi := range resp.PendingValidators {
		join.Board = append(join.Board, vi)
		join.Approved = append(join.Approved, false) // pending
	}
	return join
}

func (c *Client) CurrentValidators(ctx context.Context) ([]*vmgr.Validator, error) {
	req := &txpb.CurrentValidatorsRequest{}
	resp, err := c.txClient.CurrentValidators(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve current validators: %w", err)
	}
	vals := make([]*vmgr.Validator, len(resp.Validators))
	for i, vi := range resp.Validators {
		vals[i] = &vmgr.Validator{
			PubKey: vi.Pubkey,
			Power:  vi.Power,
		}
	}
	return vals, nil
}
