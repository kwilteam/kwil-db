package txsvc

import (
	"context"
	"encoding/json"

	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/ident"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) Call(ctx context.Context, req *txpb.CallRequest) (*txpb.CallResponse, error) {
	body, msg, err := convertActionCall(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to convert action call: %s", err.Error())
	}

	if msg.Sender != nil {
		err = ident.VerifyMessage(msg)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to verify signed message: %s", err.Error())
		}
	}

	args := make([]any, len(body.Arguments))
	for i, arg := range body.Arguments {
		args[i] = arg
	}

	executeResult, err := s.engine.Call(ctx, body.DBID, body.Action, args, msg)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to execution view action: %s", err.Error())
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

	if req.GetSignature() == nil {
		return &actionPayload, &transactions.CallMessage{
			Signature: nil,
			Body:      nil,
			Sender:    nil,
		}, nil
	}

	convSignature := conversion.ConvertToCryptoSignature(req.Signature)

	return &actionPayload, &transactions.CallMessage{
		Body: &transactions.CallMessageBody{
			Description: req.Body.Description,
			Payload:     req.Body.Payload,
		},
		Signature:     convSignature,
		Sender:        req.Sender,
		Serialization: transactions.SignedMsgSerializationType(req.Serialization),
	}, nil
}
