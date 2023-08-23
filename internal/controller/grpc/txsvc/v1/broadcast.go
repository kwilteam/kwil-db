package txsvc

import (
	"context"
	"encoding/hex"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"go.uber.org/zap"
)

func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	tx, err := convertTransaction(req.Tx)
	if err != nil {
		s.log.Warn("failed to convert transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to convert transaction: %s", err)
	}

	s.log.Debug("incoming transaction",
		zap.String("PayloadType", tx.Body.PayloadType.String()),
		zap.String("from", tx.GetSenderAddress()),
	)

	err = tx.Verify()
	if err != nil {
		s.log.Debug("failed to verify transaction", zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "failed to verify transaction: %s", err)
	}

	txHash, err := s.chainClient.BroadcastTxAsync(ctx, tx)
	if err != nil {
		s.log.Error("failed to broadcast tx", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to broadcast transaction: %s", err)
	}

	s.log.Info("broadcast transaction",
		zap.String("PayloadType", tx.Body.PayloadType.String()),
		zap.String("TxHash", hex.EncodeToString(txHash)))

	return &txpb.BroadcastResponse{
		TxHash: txHash,
	}, nil
}
