package txsvc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
	"go.uber.org/zap"
)

func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	tx, err := convertTx(req.Tx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert transaction: %s", err)
	}

	err = tx.Verify()
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to verify transaction: %s", err)
	}

	hash, err := s.chainClient.BroadcastTxAsync(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction with error:  %s", err)
	}

	s.log.Info("broadcasted transaction ", zap.String("payload_type", tx.PayloadType.String()))
	return &txpb.BroadcastResponse{
		Receipt: &txpb.TxReceipt{
			TxHash: hash,
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

func convertTx(incoming *txpb.Tx) (*kTx.Transaction, error) {
	payloadType := kTx.PayloadType(incoming.PayloadType)
	if err := payloadType.IsValid(); err != nil {
		return nil, err
	}

	convSignature, err := convertSignature(incoming.Signature)
	if err != nil {
		return nil, err
	}

	return &kTx.Transaction{
		Hash:        incoming.Hash,
		PayloadType: payloadType,
		Payload:     incoming.Payload,
		Fee:         incoming.Fee,
		Nonce:       incoming.Nonce,
		Signature:   convSignature,
		Sender:      incoming.Sender,
	}, nil
}

func newEmptySignature() (bytes []byte, sigType crypto.SignatureType) {
	return []byte{}, crypto.PK_SECP256K1_UNCOMPRESSED
}

func convertSignature(sig *txpb.Signature) (*crypto.Signature, error) {
	if sig == nil {
		emptyBts, emptyType := newEmptySignature()
		return &crypto.Signature{
			Signature: emptyBts,
			Type:      emptyType,
		}, nil
	}

	sigType := crypto.SignatureType(sig.SignatureType)
	if err := sigType.IsValid(); err != nil {
		return nil, err
	}

	return &crypto.Signature{
		Signature: sig.SignatureBytes,
		Type:      sigType,
	}, nil
}
