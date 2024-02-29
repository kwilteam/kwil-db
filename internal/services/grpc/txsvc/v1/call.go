package txsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/common"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/ident"
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
	for i, isNil := range body.NilArg { // length validation in convertActionCall
		if isNil {
			args[i] = nil
		}
	}

	tx, err := s.db.BeginReadTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	signer := msg.Sender
	caller := "" // string representation of sender, if signed.  Otherwise, empty string
	if signer != nil && msg.AuthType != "" {
		caller, err = ident.Identifier(msg.AuthType, signer)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to get caller: %s", err.Error())
		}
	}

	executeResult, err := s.engine.Call(ctx, tx, &common.ExecutionData{
		Dataset:   body.DBID,
		Procedure: body.Action,
		Args:      args,
		Signer:    signer,
		Caller:    caller,
	})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to execute view action: %s", err.Error())
	}

	// marshalling the map is less efficient, but necessary for backwards compatibility

	btsResult, err := json.Marshal(ResultMap(executeResult))
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

	if nArg := len(actionPayload.NilArg); nArg > 0 && nArg != len(actionPayload.Arguments) {
		return nil, nil, fmt.Errorf("input arguments of length %d but nil args of length %d",
			len(actionPayload.Arguments), len(actionPayload.NilArg))
	}

	return &actionPayload, &transactions.CallMessage{
		Body: &transactions.CallMessageBody{
			Payload: req.Body.Payload,
		},
		AuthType: req.AuthType,
		Sender:   req.Sender,
	}, nil
}
