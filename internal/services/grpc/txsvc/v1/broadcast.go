package txsvc

import (
	"context"
	"encoding/hex"

	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	"go.uber.org/zap"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

func (s *Service) ChainInfo(ctx context.Context, req *txpb.ChainInfoRequest) (*txpb.ChainInfoResponse, error) {
	status, err := s.chainClient.Status(ctx)
	if err != nil {
		return nil, err
	}
	return &txpb.ChainInfoResponse{
		ChainId: status.Node.ChainID,
		Height:  uint64(status.Sync.BestBlockHeight),
		Hash:    status.Sync.BestBlockHash,
	}, nil
}

func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	logger := s.log.With(zap.String("rpc", "Broadcast"),
		zap.String("PayloadType", req.Tx.Body.PayloadType))
	logger.Debug("incoming transaction")

	tx, err := conversion.ConvertFromPBTx(req.Tx)
	if err != nil {
		logger.Error("failed to convert transaction", zap.Error(err))
		// NOTE: for internal error, we should not expose the error message to the client
		return nil, status.Errorf(codes.Internal, "failed to convert transaction")
	}

	logger = logger.With(zap.String("from", hex.EncodeToString(tx.Sender)))

	encodedTx, err := tx.MarshalBinary()
	if err != nil {
		logger.Error("failed to serialize transaction data", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to serialize transaction data")
	}

	const sync = 1
	code, txHash, err := s.chainClient.BroadcastTx(ctx, encodedTx, sync)
	if err != nil {
		logger.Error("failed to broadcast tx", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to broadcast transaction")
	}

	if txCode := transactions.TxCode(code); txCode != transactions.CodeOk {
		stat := &spb.Status{
			Code:    int32(codes.InvalidArgument),
			Message: "broadcast error",
		}
		if details, err := anypb.New(&txpb.BroadcastErrorDetails{
			Code:    code, // e.g. invalid nonce, wrong chain, etc.
			Hash:    hex.EncodeToString(txHash),
			Message: txCode.String(),
		}); err != nil {
			logger.Error("failed to marshal broadcast error details", zap.Error(err))
		} else {
			logger.Info("broadcast error details", zap.Uint32("code", code), zap.String("message", txCode.String()))
			stat.Details = append(stat.Details, details)
		}
		return nil, status.ErrorProto(stat)
	}

	logger.Info("broadcast transaction", zap.String("TxHash", hex.EncodeToString(txHash)),
		zap.Int("sync", sync), zap.Uint64("nonce", tx.Body.Nonce))
	return &txpb.BroadcastResponse{
		TxHash: txHash,
	}, nil
}
