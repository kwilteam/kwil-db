package txsvc

import (
	"context"
	"encoding/json"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Call(ctx context.Context, req *txpb.CallRequest) (*txpb.CallResponse, error) {
	body, msg, err := convertActionCall(req)
	if err != nil {
		// NOTE: http api needs to be able to get the error message
		return nil, status.Errorf(codes.InvalidArgument, "failed to convert action call: %s", err.Error())
	}

	args := make([]any, len(body.Arguments))
	for i, arg := range body.Arguments {
		args[i] = arg
	}

	executeResult, err := s.engine.Call(ctx, body.DBID, body.Action, args, msg)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to execute view action: %s", err.Error())
	}

	btsResult, err := json.Marshal(executeResult)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal call result")
	}

	return &txpb.CallResponse{
		Result: btsResult,
	}, nil
}

func convertActionCall(req *txpb.CallRequest) (*transactions.ActionCall, *transactions.CallMessage, error) {
	var actionPayload transactions.ActionCall

	err := actionPayload.UnmarshalBinary(req.Body.Payload)
	if err != nil {
		return nil, nil, err
	}

	return &actionPayload, &transactions.CallMessage{
		Body: &transactions.CallMessageBody{
			Payload: req.Body.Payload,
		},
		AuthType: req.AuthType,
		Sender:   req.Sender,
	}, nil
}
