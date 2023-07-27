package txsvc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
)

func (s *Service) Call(ctx context.Context, req *txpb.CallRequest) (*txpb.CallResponse, error) {
	execBody := convertActionCall(req.Payload)
	// NOTE: if sender is used, must be validated with signature
	//sender := req.GetSender()
	//signature := req.GetSignature()

	// @yaiba Note: should only execute only if the action is `view` action
	executeResult, err := s.executor.Execute(ctx, &entity.ExecuteAction{
		Tx:            nil,
		ExecutionBody: execBody,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execution action: %w", err)
	}

	return &txpb.CallResponse{
		Result: executeResult.Body,
	}, nil
}

func convertActionCall(payload *txpb.ActionPayload) *entity.ActionExecution {
	exec := &entity.ActionExecution{
		DBID:   payload.GetDbid(),
		Action: payload.GetAction(),
		Params: make([]map[string]any, len(payload.GetParams())),
	}

	for i, input := range payload.GetParams() {
		exec.Params[i] = make(map[string]any)
		for k, v := range input.GetInput() {
			exec.Params[i][k] = v
		}
	}

	return exec
}
