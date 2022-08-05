package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/tidwall/wal"
)

// Struct for write ahead log. Contains fields for the block that the log is for
type Wal struct {
	wal  *wal.Log
	name string
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
	wal_FOLDER    string = ".kwal"
	wal_BEGIN     string = "BEGIN"
	wal_END_BLOCK string = "END_BLOCK"
)

//TODO: add recovery logic or a registerable recovery handler

var wip_FOLDER string = "" //set during ensureInitialized()

var ErrorFileNotClosed = errors.New("file must be closed in order to move it")

// Will create a new WAL based on context.
func NewBlockWal(ctx BlockContext) (Wal, error) {
	return NewBlockWalWithOptions(ctx, nil)
}

func NewBlockWalWithOptions(ctx BlockContext, walOpts *wal.Options) (Wal, error) {
	ensureInitialized()

	height := ctx.BlockHeight()
	if height < 0 {
		panic("BlockContext::BlockHeight must be >= 0")
	}

	name := leftZeroPad(uint64(height))

	// Building the string for the path
	path := path.Join(wip_FOLDER, name)

	// Opening wal
	innerWal, err := wal.Open(path, walOpts)
	if err != nil {
		return Wal{}, err
	}

	// Creating new wal
	newWal := Wal{wal: innerWal, name: name}

	// Write begin block
	newWal.appendWriteString(wal_BEGIN)

	return newWal, nil
}

// Function to finish a wal and send it to the final directory.
func (w *Wal) Seal() error {
	// Write EndBlock
	err := w.appendWriteString(wal_END_BLOCK)
	if err != nil {
		return err
	}

	// Close
	err = w.wal.Close()
	if err != nil {
		return err
	}

	// Move the location
	return w.moveSealedLog()
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

// Function to move the location of the wal. Will automatically create the new directory.
func (w *Wal) moveSealedLog() error {
	// Get current WIP file path
	source := path.Join(wip_FOLDER, w.name)

	// Get WAL file path
	target := path.Join(wal_FOLDER, w.name)

	// Rename. This does not delete the old directory
	err := os.Rename(source, target)
	if err != nil {
		return err
	}

	return nil
}

func newLogPrefix(mByte uint8, msgType uint8) []byte {
	var m []byte
	m = append(m, mByte)
	m = append(m, msgType)
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

func leftZeroPad(number uint64) string {
	return fmt.Sprintf("%020d", number)
}

// This method is idempotent and does not need to have any concurrent checkign logic, etc
func ensureInitialized() {
	if wip_FOLDER != "" {
		return
	}

	var wip = path.Join(wal_FOLDER, "_wip")

	// Making the directory
	err := os.MkdirAll(wal_FOLDER, 0755) // Is 0755 the correct FileMode for this? It is more secure, however if there is an issue with another process being unable to delete the logs, then change it to 0777
	if err != nil {
		err = os.MkdirAll(wip, 0755)
	}

	if err != nil {
		panic("unable to initialize WAL data directories")
	}

	wip_FOLDER = wip
}

/*
	Below is the schema for all logs written to the WAL
	Examples are not in bytes, but instead as uints, strings, and bools.  The different parts are delimited by a "|" (only in the examples)

	PREFIX: [magic-byte][msg-type]
	SIZE:   [1]         [1]

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
			1: Message Type
			2-65: DB ID
			66-67: Message Size
			68+: Message

	TYPE 1: DDL
		HEADER: [db-id][msg-size]
		SIZE:   [64]   [2]

		BODY: [msg]
		SIZE: [msg-size]

		FULL: [magic-byte][msg-type][db-id][msg-size][msg]
		EXAMPLE: 0|0|b7387487f514877209a87e502d2c1817669f21fac9153941292adcf995c5275e|75|75-byte-DDL-msg...

		BYTES:
			0: Magic Byte
			1: Message Type
			2-65: DB ID
			66-67: Message Size
			68+: Message

	TYPE 2: DefineQuery
		HEADER: [db-id][publicity][msg-size]
		SIZE:	[64]   [1]        [2]

		BODY: [msg]
		SIZE: [msg-size]

		FULL: [magic-byte][msg-type][db-id][publicity][msg-size][msg]
		EXAMPLE: 0|0|b7387487f514877209a87e502d2c1817669f21fac9153941292adcf995c5275e|true|75|75-byte-DDL-msg...

		BYTES:
			0: Magic Byte
			1: Message Type
			2-65: DB ID
			66: Publicity
			67-68: Message Size
			69+: Message

	TYPE 3:
		HEADER: [dbid][statementid][#-of-inputs]
		SIZE:   [64]  [64]         [1]

		BODY: [input-n-size][input-n-type][input-n+1-size][input-n+1-type]... [input-n]     [input-n+1]...
		SIZE  [2]           [1]           [2]             [1]	              [input-n-size][input-n+1-size]


		FULL: [magic-byte][msg-type][db-msg-size][dbid][statementid][BODY]
		EXAMPLE: 0|3|355|b7387487f514877209a87e502d2c1817669f21fac9153941292adcf995c5275e|4f66846bac3022f305f666848d5665d33f1db2305df56bbb625d329cf5d794b1|2|100|string|250|string|100-byte-long-data...|250-byte-long-data...

		BYTES:
			0: Magic Byte
			1: Message Type
			2-65: DB ID
			66-129: Statement ID
			130: # of inputs
			131+: Body


	TYPE N:
		HEADER:
		SIZE:

		BODY:
		SIZE:

		FULL:
		EXAMPLE:

		BYTES:
			0: Magic Byte
			1: Message Type

*/
