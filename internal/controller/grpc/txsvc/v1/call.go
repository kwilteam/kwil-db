package txsvc

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

func (s *Service) Call(ctx context.Context, req *txpb.CallRequest) (*txpb.CallResponse, error) {

	body, msg, err := convertActionCall(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to convert action call: %s", err.Error())
	}

	if msg.Sender != "" {
		err = msg.Verify()
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to verify signed message: %s", err.Error())
		}
	}

	executeResult, err := s.engine.Call(ctx, body, msg)
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

func convertActionCall(req *txpb.CallRequest) (*tx.CallActionPayload, *tx.SignedMessage[tx.JsonPayload], error) {
	var actionPayload *tx.CallActionPayload
	err := json.Unmarshal(req.Payload, &actionPayload)
	if err != nil {
		return nil, nil, err
	}

	convSignature, err := convertSignature(req.Signature)
	if err != nil {
		return nil, nil, err
	}

	return &tx.CallActionPayload{
			Action: actionPayload.Action,
			DBID:   actionPayload.DBID,
			Params: actionPayload.Params,
		}, &tx.SignedMessage[tx.JsonPayload]{
			Payload:   tx.JsonPayload(req.Payload),
			Signature: convSignature,
			Sender:    req.Sender,
		}, nil
}
