package txsvc

import (
	"context"
	"encoding/hex"
	"strings"

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

	tx, err := convertFromPBTx(req.Tx)
	if err != nil {
		// This is not necessarily an internal error. The transaction is from the client
		logger.Error("failed to convert transaction", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to convert transaction: %v", err)
	}

	logger = logger.With(zap.String("from", hex.EncodeToString(tx.Sender)))

	encodedTx, err := tx.MarshalBinary()
	if err != nil {
		logger.Error("failed to serialize transaction data", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to serialize transaction data")
	}

	var sync uint8 = 1 // default to sync, not async or commit
	if req.Sync != nil {
		sync = uint8(*req.Sync)
	}
	var commitFail bool
	res, err := s.chainClient.BroadcastTx(ctx, encodedTx, sync)
	if err != nil {
		logger.Error("failed to broadcast tx", zap.Error(err))
		if res == nil { // they really do this to report hash on commit fail/timeout
			return nil, status.Errorf(codes.Unknown, "failed to broadcast transaction: %v", err)
		} // else we have a result, and error is probably timeout
		commitFail = true // we have res, but also treat as error.
	}
	code, txHash := res.Code, res.Hash.Bytes()

	if txCode := transactions.TxCode(code); txCode != transactions.CodeOk || commitFail {
		stat := &spb.Status{
			Code:    int32(codes.InvalidArgument),
			Message: "broadcast error",
		}
		if commitFail { // we have both res and err, probably a timeout
			stat.Message = err.Error()
			if strings.Contains(err.Error(), "timed out") { // not exported; doing our best
				stat.Code = int32(codes.DeadlineExceeded)
			} else {
				stat.Code = int32(codes.Unknown)
			}
		}
		if details, err := anypb.New(&txpb.BroadcastErrorDetails{
			Code:    code, // e.g. invalid nonce, wrong chain, etc. or maybe OK if commit timed out
			Hash:    hex.EncodeToString(txHash),
			Message: res.Log,
		}); err != nil {
			logger.Error("failed to marshal broadcast error details", zap.Error(err))
		} else {
			logger.Info("broadcast error details", zap.Uint32("code", code), zap.String("message", res.Log))
			stat.Details = append(stat.Details, details)
		}
		return nil, status.ErrorProto(stat)
	}

	logger.Info("broadcast transaction", zap.String("TxHash", hex.EncodeToString(txHash)),
		zap.Uint8("sync", sync), zap.Uint64("nonce", tx.Body.Nonce))
	return &txpb.BroadcastResponse{
		TxHash: txHash,
	}, nil
}
