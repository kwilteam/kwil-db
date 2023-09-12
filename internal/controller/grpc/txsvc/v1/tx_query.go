package txsvc

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/kwilteam/kwil-db/pkg/abci"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/transactions"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) TxQuery(ctx context.Context, req *txpb.TxQueryRequest) (*txpb.TxQueryResponse, error) {
	logger := s.log.With(zap.String("rpc", "TxQuery"),
		zap.String("TxHash", strings.ToUpper(hex.EncodeToString(req.TxHash))))
	logger.Debug("query transaction")

	cmtResult, err := s.chainClient.TxQuery(ctx, req.TxHash, false)
	if err != nil {
		if errors.Is(err, abci.ErrTxNotFound) {
			logger.Debug("transaction not found")
			return nil, status.Error(codes.NotFound, "transaction not found")
		}
		logger.Error("failed to query tx", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to query transaction")
	}

	if cmtResult.Height < 0 {
		logger.Debug("transaction not found",
			zap.ByteString("TxHash", req.TxHash), zap.Int64("Height", cmtResult.Height))
		return nil, status.Error(codes.NotFound, "transaction not found")
	}

	originalTx := &transactions.Transaction{}
	if err := originalTx.UnmarshalBinary(cmtResult.Tx); err != nil {
		logger.Error("failed to deserialize transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to deserialize transaction")
	}

	tx, err := convertFromAbciTx(originalTx)
	if err != nil {
		logger.Warn("failed to convert transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to convert transaction")
	}

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
