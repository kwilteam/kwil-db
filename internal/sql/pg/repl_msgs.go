package pg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"github.com/jackc/pglogrepl"
)

const (
	MessageTypePrepare          pglogrepl.MessageType = 'P' // this is not the prepare you're looking for
	MessageTypeBeginPrepare     pglogrepl.MessageType = 'b'
	MessageTypeCommitPrepared   pglogrepl.MessageType = 'K'
	MessageTypeRollbackPrepared pglogrepl.MessageType = 'r'
	MessageTypeStreamPrepare    pglogrepl.MessageType = 'p'
)

// msgTypeToString is helpful for debugging, but normally unused.
func msgTypeToString(t pglogrepl.MessageType) string { //nolint:unused
	switch t {
	case MessageTypePrepare:
		return "Prepare"
	case MessageTypeStreamPrepare:
		return "Stream Prepared"
	case MessageTypeBeginPrepare:
		return "Begin Prepared"
	case MessageTypeCommitPrepared:
		return "Commit Prepared"
	case MessageTypeRollbackPrepared:
		return "Rollback Prepared"
	default:
		return t.String()
	}
}

// parseV3 parse a logical replication message from protocol version #3
func parseV3(data []byte, inStream bool) (m pglogrepl.Message, err error) {
	msgType := pglogrepl.MessageType(data[0])

	var decoder pglogrepl.MessageDecoder // v1 and v3 have same Decode signature (stream not relevant)

	switch msgType {
	case MessageTypePrepare, MessageTypeStreamPrepare: // same encoding
		decoder = new(PrepareMessageV3)
	case MessageTypeBeginPrepare:
		decoder = new(BeginPrepareMessageV3)
	case MessageTypeCommitPrepared:
		decoder = new(CommitPreparedMessageV3)
	case MessageTypeRollbackPrepared:
		decoder = new(RollbackPreparedMessageV3)
	default:
		return pglogrepl.ParseV2(data, inStream)
	}

	if v2, ok := decoder.(pglogrepl.MessageDecoderV2); ok {
		if err = v2.DecodeV2(data[1:], inStream); err != nil {
			return nil, err
		}
	} else if err = decoder.Decode(data[1:]); err != nil {
		return nil, err
	}

	return decoder.(pglogrepl.Message), nil
}

func decodeUint32(src []byte) (uint32, int) {
	return binary.BigEndian.Uint32(src), 4
}

func decodeLSN(src []byte) (pglogrepl.LSN, int) {
	return pglogrepl.LSN(binary.BigEndian.Uint64(src)), 8
}

const microsecFromUnixEpochToY2K = 946684800 * 1000000

func pgTimeToTime(microsecSinceY2K int64) time.Time {
	microsecSinceUnixEpoch := microsecFromUnixEpochToY2K + microsecSinceY2K
	return time.Unix(0, microsecSinceUnixEpoch*1000)
}

func decodeTime(src []byte) (time.Time, int) {
	return pgTimeToTime(int64(binary.BigEndian.Uint64(src))), 8
}

// NOTE: there is a lot of overlap in these prepared transaction structs,
// perhaps a common base?

// PrepareMessageV3 is the a prepared transaction message.
type PrepareMessageV3 struct {
	// Flags currently unused (must be 0).
	Flags uint8
	// PrepareLSN is the LSN of the prepare.
	PrepareLSN pglogrepl.LSN
	// EndPrepareLSN is the end LSN of the prepared transaction.
	EndPrepareLSN pglogrepl.LSN
	// PrepareTime is the prepare timestamp of the transaction
	PrepareTime time.Time
	// Xid of the transaction
	Xid uint32
	// UserGID ius the user-defined GID of the prepared transaction.
	UserGID string
}

func fromCString(b []byte) string {
	return string(bytes.TrimRight(b, "\x00"))
}

func (m *PrepareMessageV3) Decode(src []byte) error {
	if len(src) < 25 {
		return errors.New("too short")
	}

	var low, used int
	m.Flags = src[0]
	low++
	m.PrepareLSN, used = decodeLSN(src[low:])
	low += used
	m.EndPrepareLSN, used = decodeLSN(src[low:])
	low += used
	m.PrepareTime, used = decodeTime(src[low:])
	low += used

	m.Xid, used = decodeUint32(src)
	low += used

	m.UserGID = fromCString(src[low:])

	return nil
}

func (m *PrepareMessageV3) Type() pglogrepl.MessageType {
	return MessageTypePrepare
}

// BeginPrepareMessageV3 is the beginning of a prepared transaction message.
type BeginPrepareMessageV3 struct {
	// Flags currently unused (must be 0).
	// Flags uint8
	// PrepareLSN is the LSN of the prepare.
	PrepareLSN pglogrepl.LSN
	// EndPrepareLSN is the end LSN of the prepared transaction.
	EndPrepareLSN pglogrepl.LSN
	// PrepareTime is the prepare timestamp of the transaction
	PrepareTime time.Time
	// Xid of the transaction
	Xid uint32
	// UserGID ius the user-defined GID of the prepared transaction.
	UserGID string
}

func (m *BeginPrepareMessageV3) Decode(src []byte) error {
	if len(src) < 29 {
		return errors.New("too short")
	}

	var low, used int
	// m.Flags = src[0]
	// low += 1
	m.PrepareLSN, used = decodeLSN(src[low:])
	low += used
	m.EndPrepareLSN, used = decodeLSN(src[low:])
	low += used
	m.PrepareTime, used = decodeTime(src[low:])
	low += used

	m.Xid, used = decodeUint32(src)
	low += used

	m.UserGID = fromCString(src[low:])

	return nil
}

func (m *BeginPrepareMessageV3) Type() pglogrepl.MessageType {
	return MessageTypeBeginPrepare
}

// CommitPreparedMessageV3 is a commit prepared message.
type CommitPreparedMessageV3 struct {
	// Flags currently unused (must be 0).
	Flags uint8
	// CommitLSN is the LSN of the commit of the prepared transaction.
	CommitLSN pglogrepl.LSN
	// EndCommitLSN is the end LSN of the commit of the prepared transaction.
	EndCommitLSN pglogrepl.LSN
	// CommitTime is the commit timestamp of the transaction
	CommitTime time.Time
	// Xid of the transaction
	Xid uint32
	// UserGID is the user-defined GID of the prepared transaction.
	UserGID string
}

func (m *CommitPreparedMessageV3) Decode(src []byte) error {
	if len(src) < 25 {
		return errors.New("too short")
	}

	var low, used int
	m.Flags = src[0]
	low++
	m.CommitLSN, used = decodeLSN(src[low:])
	low += used
	m.EndCommitLSN, used = decodeLSN(src[low:])
	low += used
	m.CommitTime, used = decodeTime(src[low:])
	low += used

	m.Xid, used = decodeUint32(src)
	low += used

	m.UserGID = fromCString(src[low:])

	return nil
}

func (m *CommitPreparedMessageV3) Type() pglogrepl.MessageType {
	return MessageTypeCommitPrepared
}

// RollbackPreparedMessageV3 is a rollback prepared message.
type RollbackPreparedMessageV3 struct {
	// Flags currently unused (must be 0).
	Flags uint8
	// EndLSN is the end LSN of the prepared transaction.
	EndLSN pglogrepl.LSN
	// RollbackLSN is the end LSN of the rollback of the prepared transaction.
	RollbackLSN pglogrepl.LSN
	// PrepareTime is the prepare timestamp of the transaction
	PrepareTime time.Time
	// RollbackTime is the rollback timestamp of the transaction
	RollbackTime time.Time
	// Xid of the transaction
	Xid uint32
	// UserGID ius the user-defined GID of the prepared transaction.
	UserGID string
}

func (m *RollbackPreparedMessageV3) Decode(src []byte) error {
	if len(src) < 33 {
		return errors.New("too short")
	}

	var low, used int
	m.Flags = src[0]
	low++
	m.RollbackLSN, used = decodeLSN(src[low:])
	low += used
	m.EndLSN, used = decodeLSN(src[low:])
	low += used
	m.PrepareTime, used = decodeTime(src[low:])
	low += used
	m.RollbackTime, used = decodeTime(src[low:])
	low += used
	m.Xid, used = decodeUint32(src)
	low += used

	m.UserGID = fromCString(src[low:])

	return nil
}

func (m *RollbackPreparedMessageV3) Type() pglogrepl.MessageType {
	return MessageTypeRollbackPrepared
}
