package wal

import (
	"encoding/binary"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/tidwall/wal"
)

type BlockContext interface {
	BlockHeight() int64
}

// Struct for write ahead log.  Contains fields for the block that the log is for
type Wal struct {
	wal             *wal.Log
	height          uint64
	currentLocation string
}

var WalOpts *wal.Options
var CurrentWal Wal

const (
	TEMPFOLDER         string = "tmp/"
	FINISHEDLOGSFOLDER string = "finished_logs/"
)

func NewWal(path string) (Wal, error) {
	innerWal, err := wal.Open(path, WalOpts)
	if err != nil {
		return Wal{}, err
	}
	return Wal{
		wal:             innerWal,
		currentLocation: path,
	}, err
}

// Write any arbitrary data to the WAL at any index
func (w *Wal) Write(index uint64, data []byte) error {
	return w.wal.Write(index, data)
}

// Will automatically add the data to the end of the log
func (w *Wal) AppendWrite(data []byte) error {

	// Find the last index
	currentIndex, err := w.wal.LastIndex()
	if err != nil {
		return err
	}

	// Increment by one
	currentIndex++
	// Write
	return w.Write(currentIndex, data)
}

var ErrorFileNotClosed = errors.New("file must be closed in order to move it")

// Function to move the location of the wal.  Will automatically create the new directory.
func (w *Wal) MoveLocation(dst, fileName string) error {

	// Making the directory
	err := os.MkdirAll(dst, 0755) // Is 0755 the correct FileMode for this?  It is more secure, however if there is an issue with another process being unable to delete the logs, then change it to 0777
	if err != nil {
		return err
	}

	// Find final path
	finDst := Concat([]string{dst, fileName})

	// Create the new src
	src := Concat([]string{w.currentLocation, "/00000000000000000001"})

	// Change current location to new location
	w.currentLocation = src

	// Rename.  This does not delete the old directory
	return os.Rename(src, finDst)
}

// Function to finish a wal and send it to the final directory.
func (w *Wal) Seal() error {
	// Write EndBlock
	err := w.AppendWrite([]byte("EndBlock"))
	if err != nil {
		return err
	}

	// Close
	err = w.Close()
	if err != nil {
		return err
	}

	// Create new dst
	newDst := Concat([]string{FINISHEDLOGSFOLDER, strconv.FormatInt(int64(w.height), 10)})

	// Copy old location
	oldLoc := w.currentLocation

	// Move the location
	err = w.MoveLocation(newDst, "/00000000000000000001")
	if err != nil {
		return err
	}

	return os.RemoveAll(oldLoc)
}

// Will close the wal and move the file to the proper location
func (w *Wal) Close() error {
	return w.wal.Close()
}

// Will create a new WAL based on context.
func NewBlockWal(ctx BlockContext) (Wal, error) {
	height := ctx.BlockHeight()

	// Building the string for the path
	path := Concat([]string{TEMPFOLDER, strconv.FormatInt(height, 10)})

	// Opening wal
	innerWal, err := wal.Open(path, WalOpts)
	if err != nil {
		return Wal{}, err
	}

	// Creating new wal
	newWal := Wal{wal: innerWal, height: uint64(height), currentLocation: path}

	// Write begin block
	newWal.AppendWrite([]byte("Begin"))

	return newWal, err
}

func Concat(strArr []string) string {
	var sb strings.Builder
	for i := 0; i < len(strArr); i++ {
		sb.WriteString(strArr[i])
	}
	return sb.String()
}

func NewLogPrefix(mByte uint8, msgType uint8) []byte {
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

// Function to append a CreateDatabase message to the WAL
func (w *Wal) AppendCreateDatabase(dbid, msg string) error {
	if len(dbid) != 64 {
		return errors.New("invalid dbid")
	}
	// Construct the log entry to be appended
	m := NewLogPrefix(0, 0)

	// Append the dbid
	m = append(m, dbid...)

	// Append the message length
	m = appendStringLength(m, msg)

	// Append the message
	m = append(m, msg...)

	// Write the log entry
	return w.AppendWrite(m)
}

// Appending DDL to the WAL
// This is currently the same as CreateDatabase (besides the message ID).  I kept them separate so we can change them later if we need to.
func (w *Wal) AppendDDL(dbid, msg string) error {
	if len(dbid) != 64 {
		return errors.New("invalid dbid")
	}
	// Construct the log entry to be appended
	m := NewLogPrefix(0, 1)

	// Append the dbid
	m = append(m, dbid...)

	// Append the message length
	m = appendStringLength(m, msg)

	// Append the message
	m = append(m, msg...)

	// Write the log entry
	return w.AppendWrite(m)
}

// Appends a parameterized query definition to the WAL
// I made publicity an int8 so that it can be future compatible with the addition of more parameters.
func (w *Wal) AppendDefineQuery(dbid, msg string, publicity uint8) error {
	if len(dbid) != 64 {
		return errors.New("invalid dbid")
	}
	// Construct the log entry to be appended
	m := NewLogPrefix(0, 2)

	// Append the dbid
	m = append(m, dbid...)

	// Append the publicity
	m = append(m, publicity)

	// Append the message length
	m = appendStringLength(m, msg)

	// Append the message
	m = append(m, msg...)

	// Write the log entry
	return w.AppendWrite(m)
}

type QueryArg struct {
	argType string
	arg     []byte
}

type QueryArgs []QueryArg

func (w *Wal) AppendExecuteQuery(dbid, statementid string, args QueryArgs) error {
	if len(dbid) != 64 {
		return errors.New("invalid dbid")
	}
	if len(statementid) != 64 {
		return errors.New("invalid statementid")
	}
	// Construct the log entry to be appended
	m := NewLogPrefix(0, 3)

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
	return w.AppendWrite(m)
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
