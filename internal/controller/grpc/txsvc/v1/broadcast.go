package txsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/crypto/tmhash"
	localClient "github.com/cometbft/cometbft/rpc/client/local"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
	"go.uber.org/zap"
)

func (s *Service) Broadcast(ctx context.Context, req *txpb.BroadcastRequest) (*txpb.BroadcastResponse, error) {
	tx, err := convertTx(req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	bts, err := json.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction data: %w", err)
	}
	hash := tmhash.Sum(bts)
	fmt.Printf("Broadcasting transaction with hash %x\n", hash)
	bcClient := localClient.New(s.BcNode)
	_, err = bcClient.BroadcastTxAsync(ctx, bts)
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

	var sigBytes []byte
	var sigType crypto.SignatureType

	if incoming.Signature == nil {
		sigBytes, sigType = newEmptySignature()
	} else {
		sigBytes = incoming.Signature.SignatureBytes
		sigType = crypto.SignatureType(incoming.Signature.SignatureType)
	}

	if err := sigType.IsValid(); err != nil {
		return nil, err
	}

	return &kTx.Transaction{
		Hash:        incoming.Hash,
		PayloadType: payloadType,
		Payload:     incoming.Payload,
		Fee:         incoming.Fee,
		Nonce:       incoming.Nonce,
		Signature: crypto.Signature{
			Signature: sigBytes,
			Type:      sigType,
		},
		Sender: incoming.Sender,
	}, nil
}

func newEmptySignature() (bytes []byte, sigType crypto.SignatureType) {
	return []byte{}, crypto.PK_SECP256K1_UNCOMPRESSED
}
