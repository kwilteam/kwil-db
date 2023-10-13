package txsvc

import (
	"bytes"
	"context"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	vmgr "github.com/kwilteam/kwil-db/internal/validators"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) CurrentValidators(ctx context.Context, _ *txpb.CurrentValidatorsRequest) (*txpb.CurrentValidatorsResponse, error) {
	vals, err := s.vstore.CurrentValidators(ctx)
	if err != nil {
		s.log.Error("failed to retrieve current validators", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve current validators")
	}

	pbValidators := make([]*txpb.Validator, len(vals))
	for i, vi := range vals {
		pbValidators[i] = &txpb.Validator{
			Pubkey: vi.PubKey,
			Power:  vi.Power,
		}
	}

	return &txpb.CurrentValidatorsResponse{
		Validators: pbValidators,
	}, nil
}

func convertJoinRequest(join *vmgr.JoinRequest) *txpb.ValidatorJoinStatusResponse {
	resp := &txpb.ValidatorJoinStatusResponse{
		Power: join.Power,
	}
	for i, approved := range join.Approved {
		val := join.Board[i]
		if approved {
			resp.ApprovedValidators = append(resp.ApprovedValidators, val)
		} else {
			resp.PendingValidators = append(resp.PendingValidators, val)
		}
	}
	return resp
}

func (s *Service) ValidatorJoinStatus(ctx context.Context, req *txpb.ValidatorJoinStatusRequest) (*txpb.ValidatorJoinStatusResponse, error) {
	joiner := req.Pubkey
	allJoins, err := s.vstore.ActiveVotes(ctx)
	if err != nil {
		s.log.Error("failed to retrieve active join requests", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve active join requests")
	}
	for _, ji := range allJoins {
		if bytes.Equal(ji.Candidate, joiner) {
			return convertJoinRequest(ji), nil
		}
	}

	vals, err := s.vstore.CurrentValidators(ctx)
	if err != nil {
		s.log.Error("failed to retrieve current validators", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to retrieve current validators")
	}
	for _, vi := range vals {
		if bytes.Equal(vi.PubKey, joiner) {
			return nil, status.Errorf(codes.NotFound, "already a validator") // maybe FailedPrecondition?
		}
	}

	return nil, status.Errorf(codes.NotFound, "no active join request")
}
