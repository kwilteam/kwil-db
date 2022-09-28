package utils

import (
	"errors"
	"math/big"

	u "kwil/x/utils"
)

type WalEthTx struct {
	inner *walWriter
}

// This file contains some abstractions for writing WAL data for Ethereum event listeners

// OpenEthTxWal Will create a new WAL based on context.
func OpenEthTxWal(path string) (*WalEthTx, error) {
	inner, err := openWalWriter(path, "wal-etx")
	if err != nil {
		return nil, err
	}

	// Creating new wal
	return &WalEthTx{inner}, nil
}

func (w *WalEthTx) BeginEthBlock(h *big.Int) error {
	return w.appendEthBlock(500, h)
}

func (w *WalEthTx) EndEthBlock(h *big.Int) error {
	return w.appendEthBlock(501, h)
}

func (w *WalEthTx) BeginTransaction(tx []byte) error {
	return w.appendTransaction(502, tx)
}

func (w *WalEthTx) EndTransaction(tx []byte) error {
	return w.appendTransaction(503, tx)
}

func (w *WalEthTx) Close() {
	_ = w.inner.closeWal()
}

func (w *WalEthTx) appendEthBlock(msgType uint16, h *big.Int) error {
	m := newWalMessage(msgType).append(u.BigInt2Bytes(h)...)
	return w.inner.appendMsgToWal(m)
}

func (w *WalEthTx) appendTransaction(msgType uint16, tx []byte) error {
	if len(tx) != 32 {
		return errors.New("invalid tx hash: hash must be 32 bytes")
	}

	m := newWalMessage(msgType).append(tx[:]...)
	return w.inner.appendMsgToWal(m)
}

/*
	PREFIX: [magic-byte][msg-type]
	SIZE:   [1]         [2]

	TYPES:
		500 - Begin Block
		501 - End Block
		502 - Begin Transaction
		503 - End Transaction

	TYPE 500: Begin Block
		HEADER: [block-height]
		SIZE:   [16]

		BODY:
		SIZE:

		FULL: [magic-byte][msg-type][block-height]
		EXAMPLE: 0|500|234564323456

		BYTES:
			0: Magic Byte
			1-2: Message Type
			3-18: Block Height

	TYPE 501: End Block
		HEADER: [block-height]
		SIZE:   [16]

		BODY:
		SIZE:

		FULL: [magic-byte][msg-type][block-height]
		EXAMPLE: 0|500|234564323456

		BYTES:
			0: Magic Byte
			1-2: Message Type
			3-18: Block Height

	TYPE 502: Begin Transaction
		HEADER: [tx-hash]
		SIZE:   [32]

		BODY:
		SIZE:

		FULL: [magic-byte][msg-type][tx-hash]
		EXAMPLE: 0|502|0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef

		BYTES:
			0: Magic Byte
			1-2: Message Type
			3-34: Transaction Hash

	TYPE 503: End Transaction
		HEADER: [tx-hash]
		SIZE:   [32]

		BODY:
		SIZE:

		FULL: [magic-byte][msg-type][tx-hash]
		EXAMPLE: 0|502|0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef

		BYTES:
			0: Magic Byte
			1-2: Message Type
			3-34: Transaction Hash
*/
