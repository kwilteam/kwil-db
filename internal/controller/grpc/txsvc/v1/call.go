package txsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

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

func convertActionCall(req *txpb.CallRequest) (*entity.ActionCall, error) {
	convertedParams := make(map[string]interface{})
	for k, v := range req.Payload.Params {
		var anyVal any

		switch realVal := v.Value.(type) {
		case *txpb.ScalarValue_IntValue:
			anyVal = realVal.IntValue
		case *txpb.ScalarValue_StringValue:
			anyVal = realVal.StringValue
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown value type '%s' for param %s", reflect.TypeOf(v).String(), k)
		}

		convertedParams[k] = anyVal
	}

	convSignature, err := convertSignature(req.Signature)
	if err != nil {
		return nil, err
	}

	exec := &entity.ActionCall{
		Message: &tx.SignedMessage[*tx.CallActionPayload]{
			Payload: &tx.CallActionPayload{
				Action: req.Payload.Action,
				DBID:   req.Payload.Dbid,
				Params: convertedParams,
			},
			Signature: convSignature,
			Sender:    req.Sender,
		},
	}

	return exec, nil
}
