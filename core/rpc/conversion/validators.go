package conversion

import (
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types"
)

func ConvertFromPBJoinRequest(resp *txpb.ValidatorJoinStatusResponse) *types.JoinRequest {
	total := len(resp.ApprovedValidators) + len(resp.PendingValidators)
	join := &types.JoinRequest{
		Power:    resp.Power,
		Board:    make([][]byte, 0, total),
		Approved: make([]bool, 0, total),
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
