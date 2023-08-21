package sessions

import (
	"bytes"
	"context"
	"fmt"
)

// sessionWal is a WAL for a session
type sessionWal struct {
	wal Wal
}

// WriteBegin writes a begin record to the WAL.  This should be called before any changes are written to the WAL.
func (w *sessionWal) WriteBegin(ctx context.Context) error {
	bts, err := SerializeWalRecord(&WalRecord{
		Type: WalRecordTypeBegin,
	})
	if err != nil {
		return err
	}

	return w.wal.Append(ctx, bts)
}

// WriteCommit writes a commit record to the WAL.
// This should be called after all changes have been written to the WAL.
func (w *sessionWal) WriteCommit(ctx context.Context) error {
	bts, err := SerializeWalRecord(&WalRecord{
		Type: WalRecordTypeCommit,
	})
	if err != nil {
		return err
	}

	return w.wal.Append(ctx, bts)
}

// WriteChangeset writes a changeset for a specific comittableId to the WAL
func (w *sessionWal) WriteChangeset(ctx context.Context, committable CommittableId, changeset []byte) error {
	bts, err := SerializeWalRecord(&WalRecord{
		Type:          WalRecordTypeChangeset,
		CommittableId: committable,
		Data:          changeset,
	})
	if err != nil {
		return err
	}

	return w.wal.Append(ctx, bts)
}

// ReadNext reads the next entry from the WAL and returns it
func (w *sessionWal) ReadNext(ctx context.Context) (*WalRecord, error) {
	data, err := w.wal.ReadNext(ctx)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("invalid wal record: %x", data)
	}

	return DeserializeWalRecord(data)
}

// Truncate truncates the WAL, deleting all entries
func (w *sessionWal) Truncate(ctx context.Context) error {
	return w.wal.Truncate(ctx)
}

type WalRecordType uint8

const (
	WalRecordTypeBegin WalRecordType = iota
	WalRecordTypeCommit
	WalRecordTypeChangeset
)

// WalRecord is a record in the WAL.
// it contains a type signaling the type of record, and the data for the record
type WalRecord struct {
	Type          WalRecordType
	CommittableId CommittableId
	Data          []byte
}

// SerializeWalRecord serializes a wal record into a byte slice
func SerializeWalRecord(record *WalRecord) ([]byte, error) {
	var buf bytes.Buffer

	// magic byte
	buf.WriteByte(0x00)

	// record type
	buf.WriteByte(byte(record.Type))

	commitableIdLen := len(record.CommittableId)
	if commitableIdLen > 255 {
		return nil, fmt.Errorf("commitableId '%s' is too long: %d", record.CommittableId, commitableIdLen)
	}

	// length of committable id
	buf.WriteByte(byte(commitableIdLen))

	// committable id
	buf.Write(record.CommittableId.Bytes())

	// data
	buf.Write(record.Data)

	return buf.Bytes(), nil
}

// DeserializeWalRecord deserializes a wal record from a byte slice
func DeserializeWalRecord(data []byte) (*WalRecord, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("invalid wal record: %x", data)
	}

	// magic byte
	if data[0] != 0x00 {
		return nil, fmt.Errorf("invalid wal record: %x", data)
	}

	// record type
	recordType := WalRecordType(data[1])
	if recordType > WalRecordTypeChangeset {
		return nil, fmt.Errorf("invalid wal record: %x", data)
	}

	// check if this is the end of the record
	if recordType == WalRecordTypeBegin || recordType == WalRecordTypeCommit {
		return &WalRecord{
			Type: recordType,
		}, nil
	}

	// length of committable id
	commitableIdLen := int(data[2])
	if len(data) < 3+commitableIdLen {
		return nil, fmt.Errorf("invalid wal record: %x", data)
	}

	return &WalRecord{
		Type:          recordType,
		CommittableId: CommittableId(data[3 : 3+commitableIdLen]),
		Data:          data[3+commitableIdLen:],
	}, nil
}
