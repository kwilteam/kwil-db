package wal

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/kwilteam/kwil-db/internal/utils"
)

var (
	ErrorBeginBlock = errors.New("BeginBlock can only be called on first call or after EndBlock operation")

	ErrorEndBlock = errors.New("out of order operation, unable to set EndBlock operation")

	//ErrorWrite = errors.New("BeginBlock must be called before an Append operation")

	//ErrorWriteGeneral = errors.New("an error occurred while writing to the WAL")

	ErrorWalClosed = errors.New("wal has already been closed")

	ErrorInvalidState = errors.New("wal is in an invalid state")

	ErrorNotWritable = errors.New("wal is not writable")
)

type QueryArg struct {
	argType string
	arg     []byte
}

type QueryArgs []QueryArg

const (
	xBEGIN_BLOCK string = "BEGIN_BLOCK(%d)"
	xEND_BLOCK   string = "(%d)END_BLOCK"
)

type WalDbCmd struct {
	inner        *walWriter
	blockStarted *uint32
}

//goland:noinspection GoUnusedExportedFunction
func OpenDbCmdWal(path string) *WalDbCmd {
	w, err := openWalWriter(path, "dbcmd.wal")
	utils.PanicIfErrorMsg(err, "unable to open WAL file.")

	v := uint32(0)

	return &WalDbCmd{w, &v}
}

func OpenDbCmdWalFromHomeDir(path string) *WalDbCmd {
	w, err := openWalWriterFromHomeDir(path, "dbcmd.wal")
	utils.PanicIfErrorMsg(err, "unable to open WAL file.")

	v := uint32(0)

	return &WalDbCmd{w, &v}
}

func (w *WalDbCmd) BeginBlock(h int64) error {
	if atomic.CompareAndSwapUint32(w.blockStarted, 0, 1) {
		return w.inner.appendRawToWal([]byte(fmt.Sprintf(xBEGIN_BLOCK, h)))
	}

	// essentially close it out since we need to error out
	w.Close()
	return ErrorBeginBlock
}

func (w *WalDbCmd) EndBlock(h int64) error {
	err := w.inner.appendRawToWal([]byte(fmt.Sprintf(xEND_BLOCK, h)))
	if err != nil {
		return err
	}

	if atomic.CompareAndSwapUint32(w.blockStarted, 1, 0) {
		return nil
	}

	w.Close()

	return ErrorEndBlock
}

func (w *WalDbCmd) IsClosed() bool {
	return uint32(86) == *w.blockStarted
}

func (w *WalDbCmd) Close() {
	if uint32(86) == *w.blockStarted {
		return
	}

	*w.blockStarted = 86

	_ = w.inner.closeWal()
}

// AppendCreateDatabase Function to append a CreateDatabase message to the WAL
func (w *WalDbCmd) AppendCreateDatabase(dbid, msg string) error {
	return w.appendMsgString(0, dbid, msg)
}

// AppendDDL Appending DDL to the WAL
// This is currently the same as CreateDatabase (besides the message ID).  I kept them separate so we can change them later if we need to.
func (w *WalDbCmd) AppendDDL(dbid, msg string) error {
	return w.appendMsgString(1, dbid, msg)
}

// AppendDefineQuery Appends a parameterized query definition to the WAL
// I made publicity an int8 so that it can be future compatible with the addition of more parameters.
func (w *WalDbCmd) AppendDefineQuery(dbid, msg string, publicity uint8) error {
	m, err := newWalDbMessage(2, dbid)
	if err != nil {
		return err
	}

	// Append the publicity
	m = m.append(publicity)

	// Append the message length and message
	m = m.appendLenWithString(msg)

	// Write the log entry
	return w.append(m)
}

func (w *WalDbCmd) AppendExecuteQuery(dbid, statementid string, args QueryArgs) error {
	m, err := newWalDbMessage(3, dbid)
	if err != nil {
		return err
	}

	if len(statementid) != 64 {
		return errors.New("invalid statementid")
	}

	// Append the statementid
	m = m.appendString(statementid)

	// Append the amount of arguments
	m = appendArgAmt(m, args)

	// I will create the body and then append the whole thing
	var b []byte

	// Looping through all args to add the size and type
	for i := 0; i < len(args); i++ {
		b = utils.AppendByteArrLength(b, args[i].arg)
		b = append(b, args[i].argType...)
	}

	// Looping through all args to add the actual data
	for i := 0; i < len(args); i++ {
		b = append(b, args[i].arg...)
	}

	// Append the body to the message
	m = m.append(b...)

	// Write the log entry
	return w.append(m)
}

func (w *WalDbCmd) appendMsgString(msgType uint16, dbid, msg string) error {
	m, err := newWalDbMessage(msgType, dbid)
	if err != nil {
		return err
	}

	// Append the message length and message
	m = m.appendLenWithString(msg)

	// Write the log entry
	return w.append(m)
}

func (w *WalDbCmd) append(m *walMessage) error {
	if !atomic.CompareAndSwapUint32(w.blockStarted, 1, 2) {
		if uint32(86) == *w.blockStarted {
			return ErrorWalClosed
		}

		w.Close()

		return ErrorNotWritable
	}

	err := w.inner.appendMsgToWal(m)
	if err != nil {
		return err
	}

	if !atomic.CompareAndSwapUint32(w.blockStarted, 2, 1) {
		w.Close()
		return ErrorInvalidState
	}

	return nil
}

func newWalDbMessage(msgType uint16, dbid string) (*walMessage, error) {
	if len(dbid) != 64 {
		return nil, errors.New("invalid dbid")
	}

	// Construct the log entry to be appended and append dbid
	return newWalMessage(msgType).appendString(dbid), nil
}

// Will append the slice length as uint16 to the end of the byte slice
// Use this function instead of doing it manually since this uses uint16 instead of int64
func appendArgAmt(w *walMessage, a QueryArgs) *walMessage {
	return w.append(utils.Uint16ToBytes(uint16(len(a)))...)
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
