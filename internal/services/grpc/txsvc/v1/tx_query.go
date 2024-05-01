package txsvc

import (
	"context"
	"encoding/hex"
	"errors"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/abci"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) TxQuery(ctx context.Context, req *txpb.TxQueryRequest) (*txpb.TxQueryResponse, error) {
	logger := s.log.With(zap.String("rpc", "TxQuery"),
		zap.String("TxHash", hex.EncodeToString(req.TxHash)))

	cmtResult, err := s.chainClient.TxQuery(ctx, req.TxHash, false)
	if err != nil {
		if errors.Is(err, abci.ErrTxNotFound) {
			logger.Debug("transaction not found")
			return nil, status.Error(codes.NotFound, "transaction not found")
		}
		logger.Warn("failed to query tx", zap.Error(err))
		return nil, status.Error(codes.Unknown, "failed to query transaction")
	}

	originalTx := &transactions.Transaction{}
	if err := originalTx.UnmarshalBinary(cmtResult.Tx); err != nil {
		logger.Error("failed to deserialize transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to deserialize transaction")
	}

	tx := convertToPBTx(originalTx)

	txResult := &txpb.TransactionResult{
		Code:      cmtResult.TxResult.Code,
		Log:       cmtResult.TxResult.Log,
		GasUsed:   cmtResult.TxResult.GasUsed,
		GasWanted: cmtResult.TxResult.GasWanted,
		//Data: cmtResult.TxResult.Data,
		//Events: cmtResult.TxResult.Events,
	}

	logger.Debug("tx query result", zap.Any("result", txResult))

	return &txpb.TxQueryResponse{
		Hash:     cmtResult.Hash.Bytes(),
		Height:   cmtResult.Height,
		Tx:       tx,
		TxResult: txResult,
	}, nil
}
