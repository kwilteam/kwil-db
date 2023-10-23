package txsvc

import (
	"context"
	"encoding/hex"

	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	tx, err := conversion.ConvertToAbciTx(req.Tx)
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

	const sync = 1 // async, TODO: sync field of BroadcastRequest
	txHash, err := s.chainClient.BroadcastTx(ctx, encodedTx, sync)
	if err != nil {
		logger.Error("failed to broadcast tx", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to broadcast transaction")
	}

	logger.Info("broadcast transaction", zap.String("TxHash", hex.EncodeToString(txHash)),
		zap.Int("sync", sync), zap.Uint64("nonce", tx.Body.Nonce))
	return &txpb.BroadcastResponse{
		TxHash: txHash,
	}, nil
}
