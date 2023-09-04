package txsvc

import (
	"context"
	"encoding/hex"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"go.uber.org/zap"
)

func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	logger := s.log.With(zap.String("rpc", "Broadcast"),
		zap.String("PayloadType", req.Tx.Body.PayloadType))
	logger.Debug("incoming transaction")

	tx, err := convertToAbciTx(req.Tx)
	if err != nil {
		logger.Error("failed to convert transaction", zap.Error(err))
		// NOTE: for internal error, we should not expose the error message to the client
		return nil, status.Errorf(codes.Internal, "failed to convert transaction")
	}

	logger = logger.With(zap.String("from", tx.GetSenderAddress()))

	err = tx.Verify()
	if err != nil {
		logger.Error("failed to verify transaction", zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "failed to verify transaction: %s", err)
	}

	encodedTx, err := tx.MarshalBinary()
	if err != nil {
		logger.Error("failed to serialize transaction data", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to serialize transaction data")
	}

	const sync = 1 // async, TODO: sync field of BroadcastRequest
	code, txHash, err := s.chainClient.BroadcastTx(ctx, encodedTx, sync)
	if err != nil {
		logger.Error("failed to broadcast tx", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to broadcast transaction")
	}

	logger.Info("broadcast transaction", zap.String("TxHash", strings.ToUpper(hex.EncodeToString(txHash))),
		zap.Uint32("code", code), zap.Int("sync", sync))
	return &txpb.BroadcastResponse{
		TxHash: txHash,
	}, nil
}
