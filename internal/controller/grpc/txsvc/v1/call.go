package txsvc

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func (s *Service) Call(ctx context.Context, req *txpb.CallRequest) (*txpb.CallResponse, error) {

	body, msg, err := convertActionCall(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to convert action call: %s", err.Error())
	}

	if msg.Sender == nil {
		err = msg.Verify()
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

func convertActionCall(req *txpb.CallRequest) (*transactions.ActionCall, *transactions.SignedMessage, error) {
	var actionPayload transactions.ActionCall
	err := actionPayload.UnmarshalBinary(req.Payload)
	if err != nil {
		return nil, nil, err
	}

	sender, err := crypto.PublicKeyFromBytes(req.Sender)
	if err != nil {
		return nil, nil, err
	}

	convSignature, err := convertSignature(req.Signature)
	if err != nil {
		return nil, nil, err
	}

	return &actionPayload, &transactions.SignedMessage{
		Message:   req.Payload,
		Signature: convSignature,
		Sender:    sender,
	}, nil
}
