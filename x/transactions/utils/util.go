package utils

import (
	"kwil/x/crypto"
	"kwil/x/proto/txpb"
	"kwil/x/transactions"
	"kwil/x/transactions/dto"
)

// an interface for tx's sent over GRPC
type TxMsg interface {
	GetHash() []byte
	GetPayloadType() txpb.PayloadType
	GetPayload() []byte
	GetFee() string
	GetNonce() int64
	GetSignature() *txpb.Signature
	GetSender() string
}

func TxFromMsg(txmsg TxMsg) *dto.Transaction {
	sig := txmsg.GetSignature()

	return &dto.Transaction{
		PayloadType: transactions.PayloadType(txmsg.GetPayloadType()),
		Hash:        txmsg.GetHash(),
		Payload:     txmsg.GetPayload(),
		Fee:         txmsg.GetFee(),
		Nonce:       txmsg.GetNonce(),
		Signature: crypto.Signature{
			Signature: sig.GetSignatureBytes(),
			Type:      crypto.SignatureType(sig.GetSignatureType()),
		},
		Sender: txmsg.GetSender(),
	}
}

func TxToMsg(tx *dto.Transaction) *txpb.Tx {
	return &txpb.Tx{
		Hash:        tx.Hash,
		PayloadType: txpb.PayloadType(tx.PayloadType),
		Payload:     tx.Payload,
		Fee:         tx.Fee,
		Nonce:       tx.Nonce,
		Signature:   &txpb.Signature{SignatureBytes: tx.Signature.Signature, SignatureType: txpb.SignatureType(tx.Signature.Type)},
		Sender:      tx.Sender,
	}
}
