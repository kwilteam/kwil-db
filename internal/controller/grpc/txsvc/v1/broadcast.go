package txsvc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"go.uber.org/zap"
)

func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	tx, err := convertTransaction(req.Tx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert transaction: %s", err)
	}

	err = tx.Verify()
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to verify transaction: %s", err)
	}

	err = s.chainClient.BroadcastTxAsync(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction with error:  %s", err)
	}

	txHash, err := tx.GetHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction hash with error: %s", err)
	}

	s.log.Info("broadcasted transaction ", zap.String("payload_type", tx.Body.PayloadType.String()), zap.ByteString("ID", txHash))
	return &txpb.BroadcastResponse{
		Status: &txpb.TransactionStatus{
			Id:     txHash,
			Status: transactions.StatusPending.String(),
		},
	}, nil
}

// func handleReceipt(r *kTx.Receipt, err error) (*txpb.BroadcastResponse, error) {
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &txpb.BroadcastResponse{
// 		Receipt: &txpb.TxReceipt{
// 			TxHash: r.TxHash,
// 			Fee:    r.Fee,
// 			Body:   r.Body,
// 		},
// 	}, nil
// }
