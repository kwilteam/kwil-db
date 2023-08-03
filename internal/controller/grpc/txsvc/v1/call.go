package txsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

func (s *Service) Call(ctx context.Context, req *txpb.CallRequest) (*txpb.CallResponse, error) {

	execBody, err := convertActionCall(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to convert action call: %s", err.Error())
	}

	if execBody.Message.Sender != "" {
		fmt.Println(execBody.Message.Sender)
		err = execBody.Message.Verify()
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to verify signed message: %s", err.Error())
		}
	}

	executeResult, err := s.executor.Call(ctx, execBody)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execution action: %s", err.Error())
	}

	btsResult, err := json.Marshal(executeResult)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal result: %s", err.Error())
	}

	return &txpb.CallResponse{
		Result: btsResult,
	}, nil
}

func convertActionCall(req *txpb.CallRequest) (*entity.CallAction, error) {
	var actionPayload *tx.CallActionPayload
	err := json.Unmarshal(req.Payload, &actionPayload)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal payload: %s", err.Error())
	}

	convSignature, err := convertSignature(req.Signature)
	if err != nil {
		return nil, err
	}

	exec := &entity.CallAction{
		Message: &tx.SignedMessage[tx.JsonPayload]{
			Payload:   tx.JsonPayload(req.Payload),
			Signature: convSignature,
			Sender:    req.Sender,
		},
		Payload: &tx.CallActionPayload{
			Action: actionPayload.Action,
			DBID:   actionPayload.DBID,
			Params: actionPayload.Params,
		},
	}

	return exec, nil
}
