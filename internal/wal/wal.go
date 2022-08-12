package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/tidwall/wal"
)

// Struct for write ahead log. Contains fields for the block that the log is for
type Wal struct {
	mu  sync.RWMutex
	wal *wal.Log
}

type QueryArg struct {
	argType string
	arg     []byte
}

type QueryArgs []QueryArg

type BlockContext interface {
	BlockHeight() int64
}

const (
	wal_BEGIN_BLOCK string = "BEGIN_BLOCK(%d)"
	wal_END_BLOCK   string = "END_BLOCK(%d)"
)

//TODO: add recovery logic or a registerable recovery handler

// Will create a new WAL based on context.
func Open(path string) (*Wal, error) {
	innerWal, err := wal.Open(path, nil) //need to use walOptions for reader optimization
	if err != nil {
		return nil, err
	}

	// Creating new wal
	return &Wal{wal: innerWal}, nil
}

func (w *Wal) BeginBlock(ctx BlockContext) error {
	return w.appendWriteString(fmt.Sprintf(wal_BEGIN_BLOCK, ctx.BlockHeight()))
}

func (w *Wal) EndBlock(ctx BlockContext) error {
	return w.appendWriteString(fmt.Sprintf(wal_END_BLOCK, ctx.BlockHeight()))
}

// Function to finish a wal and send it to the final directory.
func (w *Wal) Close() error {
	return w.wal.Close()
}

// Function to append a CreateDatabase message to the WAL
func (w *Wal) AppendCreateDatabase(dbid, msg string) error {
	if len(dbid) != 64 {
		return errors.New("invalid dbid")
	}
	// Construct the log entry to be appended
	m := newLogPrefix(0, 0)

	// Append the dbid
	m = append(m, dbid...)

	// Append the message length
	m = appendStringLength(m, msg)

	// Append the message
	m = append(m, msg...)

	// Write the log entry
	return w.appendWrite(m)
}

// Appending DDL to the WAL
// This is currently the same as CreateDatabase (besides the message ID).  I kept them separate so we can change them later if we need to.
func (w *Wal) AppendDDL(dbid, msg string) error {
	if len(dbid) != 64 {
		return errors.New("invalid dbid")
	}
	// Construct the log entry to be appended
	m := newLogPrefix(0, 1)

	// Append the dbid
	m = append(m, dbid...)

	// Append the message length
	m = appendStringLength(m, msg)

	// Append the message
	m = append(m, msg...)

	// Write the log entry
	return w.appendWrite(m)
}

// Appends a parameterized query definition to the WAL
// I made publicity an int8 so that it can be future compatible with the addition of more parameters.
func (w *Wal) AppendDefineQuery(dbid, msg string, publicity uint8) error {
	if len(dbid) != 64 {
		return errors.New("invalid dbid")
	}
	// Construct the log entry to be appended
	m := newLogPrefix(0, 2)

	// Append the dbid
	m = append(m, dbid...)

	// Append the publicity
	m = append(m, publicity)

	// Append the message length
	m = appendStringLength(m, msg)

	// Append the message
	m = append(m, msg...)

	// Write the log entry
	return w.appendWrite(m)
}

func (w *Wal) AppendExecuteQuery(dbid, statementid string, args QueryArgs) error {
	if len(dbid) != 64 {
		return errors.New("invalid dbid")
	}
	if len(statementid) != 64 {
		return errors.New("invalid statementid")
	}
	// Construct the log entry to be appended
	m := newLogPrefix(0, 3)

	// Append the dbid
	m = append(m, dbid...)

	// Append the statementid
	m = append(m, statementid...)

	// Append the amount of arguments
	m = appendArgAmt(m, args)

	// I will create the body and then append the whole thing
	var b []byte

	// Looping through all args to add the size and type
	for i := 0; i < len(args); i++ {
		b = appendByteArrLength(b, args[i].arg)
		b = append(b, args[i].argType...)
	}
	// Looping through all args to add the actual data
	for i := 0; i < len(args); i++ {
		b = append(b, args[i].arg...)
	}

	// Append the body to the message
	m = append(m, b...)

	// Write the log entry
	return w.appendWrite(m)
}

// Will automatically add the string data to the end of the log
func (w *Wal) appendWriteString(data string) error {
	return w.appendWrite([]byte(data))
}

// Will automatically add the data to the end of the log
func (w *Wal) appendWrite(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Find the last index
	currentIndex, err := w.wal.LastIndex()
	if err != nil {
		return err
	}

	// Increment by one
	currentIndex++
	// Write
	return w.wal.Write(currentIndex, data)
}

func newLogPrefix(mByte uint8, msgType uint16) []byte {
	var m []byte
	m = append(m, mByte)
	m = append(m, uint16ToBytes(msgType)...)
	return m
}

func uint16ToBytes(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return b
}

// Will append the string length as uint16 to the end of the byte slice
// Use this function instead of doing it manually since this uses uint16 instead of int64
func appendStringLength(b []byte, s string) []byte {
	b = append(b, uint16ToBytes(uint16(len(s)))...)
	return b
}

// Will append the slice length as uint16 to the end of the byte slice
// Use this function instead of doing it manually since this uses uint16 instead of int64
func appendArgAmt(b []byte, a QueryArgs) []byte {
	b = append(b, uint16ToBytes(uint16(len(a)))...)
	return b
}

// Will append the slice length as uint16 to the end of the byte slice
// Use this function instead of doing it manually since this uses uint16 instead of int64
func appendByteArrLength(b []byte, a []byte) []byte {
	b = append(b, uint16ToBytes(uint16(len(a)))...)
	return b
}

/*
	Below is the schema for all logs written to the WAL
	Examples are not in bytes, but instead as uints, strings, and bools.  The different parts are delimited by a "|" (only in the examples)

	PREFIX: [magic-byte][msg-type]
	SIZE:   [1]         [2]

	TYPES:
		0: CreateDatabase
		1: DDL
		2: DefineQuery
		3: DatabaseWrite

	TYPE 0: CreateDatabase
		HEADER: [db-id][msg-size]
		SIZE:   [64]   [2]

		BODY: [msg]
		SIZE: [msg-size]

		FULL: [magic-byte][msg-type][db-id][msg-size][msg]
		EXAMPLE: 0|0|b7387487f514877209a87e502d2c1817669f21fac9153941292adcf995c5275e|75|75-byte-msg...

		BYTES:
			0: Magic Byte
			1-2: Message Type
			3-66: DB ID
			67-68: Message Size
			69+: Message

	TYPE 1: DDL
		HEADER: [db-id][msg-size]
		SIZE:   [64]   [2]

		BODY: [msg]
		SIZE: [msg-size]

		FULL: [magic-byte][msg-type][db-id][msg-size][msg]
		EXAMPLE: 0|0|b7387487f514877209a87e502d2c1817669f21fac9153941292adcf995c5275e|75|75-byte-DDL-msg...

		BYTES:
			0: Magic Byte
			1-2: Message Type
			3-66: DB ID
			67-68: Message Size
			69+: Message

	TYPE 2: DefineQuery
		HEADER: [db-id][publicity][msg-size]
		SIZE:	[64]   [1]        [2]

		BODY: [msg]
		SIZE: [msg-size]

		FULL: [magic-byte][msg-type][db-id][publicity][msg-size][msg]
		EXAMPLE: 0|0|b7387487f514877209a87e502d2c1817669f21fac9153941292adcf995c5275e|true|75|75-byte-DDL-msg...

		BYTES:
			0: Magic Byte
			1-2: Message Type
			3-66: DB ID
			67: Publicity
			68-69: Message Size
			70+: Message

	TYPE 3:
		HEADER: [dbid][statementid][#-of-inputs]
		SIZE:   [64]  [64]         [1]

		BODY: [input-n-size][input-n-type][input-n+1-size][input-n+1-type]... [input-n]     [input-n+1]...
		SIZE  [2]           [1]           [2]             [1]	              [input-n-size][input-n+1-size]


		FULL: [magic-byte][msg-type][db-msg-size][dbid][statementid][BODY]
		EXAMPLE: 0|3|355|b7387487f514877209a87e502d2c1817669f21fac9153941292adcf995c5275e|4f66846bac3022f305f666848d5665d33f1db2305df56bbb625d329cf5d794b1|2|100|string|250|string|100-byte-long-data...|250-byte-long-data...

		BYTES:
			0: Magic Byte
			1-2: Message Type
			3-66: DB ID
			67-130: Statement ID
			131: # of inputs
			132+: Body


	TYPE N:
		HEADER:
		SIZE:

		BODY:
		SIZE:

		FULL:
		EXAMPLE:

		BYTES:
			0: Magic Byte
			1-2: Message Type

*/
