package txsvc

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (s *Service) TxQuery(ctx context.Context, req *txpb.TxQueryRequest) (*txpb.TxQueryResponse, error) {
	cmtResult, err := s.chainClient.TxQuery(ctx, req.TxHash, false)
	if err != nil {
		s.log.Error("failed to query tx", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to query transaction: %s", err)
	}

	if cmtResult.Height < 0 {
		s.log.Debug("transaction not found",
			zap.ByteString("TxHash", req.TxHash), zap.Int64("Height", cmtResult.Height))
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	var tx *txpb.Transaction
	if err := proto.Unmarshal(cmtResult.Tx, tx); err != nil {
		s.log.Error("failed to deserialize transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize transaction: %w", err)
	}

	txResult := &txpb.TransactionResult{
		Code:      cmtResult.TxResult.Code,
		Log:       cmtResult.TxResult.Log,
		GasUsed:   cmtResult.TxResult.GasUsed,
		GasWanted: cmtResult.TxResult.GasWanted,
		//Data: cmtResult.TxResult.Data,
	}

	return &txpb.TxQueryResponse{
		Hash:     cmtResult.Hash.Bytes(),
		Height:   uint64(cmtResult.Height),
		Tx:       tx,
		TxResult: txResult,
	}, nil
}
