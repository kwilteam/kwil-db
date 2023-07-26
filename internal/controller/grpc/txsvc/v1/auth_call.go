package txsvc

import (
	"context"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthCall handles a call request to `auth_view` action, which need to be authenticated.
// The req contains a transaction object.
func (s *Service) AuthCall(ctx context.Context, req *txpb.AuthCallRequest) (*txpb.AuthCallResponse, error) {
	tx, err := convertTx(req.Tx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to convert transaction: %w", err)
	}

	err = tx.Verify()
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to verify transaction: %w", err)
	}

	switch tx.PayloadType {
	case kTx.EXECUTE_ACTION:
		// @yaiba NOTE: should only execute only if the action is `auth_view` action
		executeResult, err := s.executeAction(ctx, tx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to call action: %w", err)
		}
		return &txpb.AuthCallResponse{Result: executeResult.Body}, nil
	default:
		return nil, status.Errorf(codes.Unimplemented, "unsupported payload type")
	}
}
