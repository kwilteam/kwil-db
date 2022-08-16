package wal

import (
	"errors"
	"math/big"
)

// This file contains some abstractions for writing WAL data for Ethereum event listeners

// This function logs that we have begun a new ETH block
func (w *Wal) BeginEthBlock(h *big.Int) error {
	m := newLogPrefix(0, 500)
	m = append(m, BigInt2Bytes(h)...)
	return w.appendWrite(m)
}

// This function logs that we have ended an ETH block
func (w *Wal) EndEthBlock(h *big.Int) error {
	m := newLogPrefix(0, 501)
	m = append(m, BigInt2Bytes(h)...)
	return w.appendWrite(m)
}

// BeginTransaction calls when we receive a new TX eth
func (w *Wal) BeginTransaction(tx []byte) error {
	if len(tx) != 32 {
		return errors.New("invalid tx hash: hash must be 32 bytes")
	}
	m := newLogPrefix(0, 502)
	m = append(m, tx[:]...)
	return w.appendWrite(m)
}

// EndTransaction calls when we have processed an eth event
func (w *Wal) EndTransaction(tx []byte) error {
	if len(tx) != 32 {
		return errors.New("invalid tx hash: hash must be 32 bytes")
	}
	m := newLogPrefix(0, 503)
	m = append(m, tx[:]...)
	return w.appendWrite(m)
}

// This function converts a big int to bytes.  The result will always be a byte slice of length 16.
func BigInt2Bytes(h *big.Int) []byte {
	b := make([]byte, 16)
	k := h.FillBytes(b)
	return k
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
