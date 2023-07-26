package txsvc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
)

func (s *Service) Call(ctx context.Context, req *txpb.CallRequest) (*txpb.CallResponse, error) {
	exec := convertActionCall(req)

	// @yaiba Note: should only execute only if the action is `view` action
	executeResult, err := s.executor.Execute(ctx, &entity.ExecuteAction{
		ExecutionBody: exec,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execution action: %w", err)
	}

	return &txpb.CallResponse{
		Result: executeResult.Body,
	}, nil
}

func convertActionCall(req *txpb.CallRequest) *entity.ActionExecution {
	exec := &entity.ActionExecution{
		DBID:   req.GetDbid(),
		Action: req.GetAction(),
		Params: make([]map[string]any, len(req.GetInputs())),
	}

	for i, input := range req.GetInputs() {
		exec.Params[i] = make(map[string]any)
		for k, v := range input.GetInput() {
			exec.Params[i][k] = v
		}
	}

	return exec
}
