package txsvc

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		// NOTE: here we can errors.Is and return an actual response with a
		// field for the error code and message instead of this internal grpc
		// thing. Since we do not have such response structure, we have to pick
		// from the general categories of gRPC response codes, which are similar
		// to http status codes, and do string matching on the message.
		if errors.Is(err, transactions.ErrWrongChain) {
			return nil, status.Errorf(codes.InvalidArgument, "wrong chain ID %q", tx.Body.ChainID)
		}
		logger.Error("failed to broadcast tx", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to broadcast transaction")
	}

	logger.Info("broadcast transaction", zap.String("TxHash", hex.EncodeToString(txHash)),
		zap.Int("sync", sync), zap.Uint64("nonce", tx.Body.Nonce))
	return &txpb.BroadcastResponse{
		TxHash: txHash,
	}, nil
}
