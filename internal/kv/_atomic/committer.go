package atomic

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

// This package is using big endian byte ordering to encode integer values such
// as data lengths. This convention should not be changed. The slice
// serialization convention defined by Kwil's SerializeSlice and
// DeserializeSlice must also be used. If either of these aspects are changed,
// the value returned by ID will be different.

func uint16ToBytes(i uint16) []byte {
	bts := make([]byte, 2)
	binary.BigEndian.PutUint16(bts, i)
	return bts
}

var bytesToUint16 = binary.BigEndian.Uint16

type kvOperation uint8

const (
	kvOperationSet kvOperation = iota
	kvOperationDelete
)

// keyValue is a basic struct containing keys and values
// it can be quickly serialized and deserialized, which is
// used for writing to and receiving data from the AtomicCommitter
type keyValue struct {
	Operation kvOperation
	// Key cannot be longer than 65535 bytes
	Key   []byte
	Value []byte // value can be any length
}

// MarshalBinary appends the length of the key, the key, and the value
func (k *keyValue) MarshalBinary() ([]byte, error) {
	var buf []byte

	// write operation
	buf = append(buf, byte(k.Operation))

	buf = append(buf, uint16ToBytes(uint16(len(k.Key)))...)
	buf = append(buf, k.Key...)

	// write value
	buf = append(buf, k.Value...)

	return buf, nil
}

// UnmarshalBinary reads the length of the key, the key, and the value
func (k *keyValue) UnmarshalBinary(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("data too short")
	}
	// read operation
	k.Operation = kvOperation(data[0])

	// read key length and key
	keyLen := bytesToUint16(data[1:3])
	if len(data) < int(keyLen)+3 {
		return fmt.Errorf("data too short")
	}
	k.Key = data[3 : keyLen+3]

	// read value
	k.Value = data[keyLen+3:]
	return nil
}

/*
	The below implements the sessions.Committable interface
*/

func (k *AtomicKV) BeginCommit(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.inSession {
		return ErrSessionActive
	}
	k.inSession = true
	return nil
}

func (k *AtomicKV) EndCommit(ctx context.Context, appender func([]byte) error) (err error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.inSession {
		return ErrSessionNotActive
	}
	k.inSession = false

	// flush uncommitted data
	bts, err := k.flushUncommittedData()
	if err != nil {
		return err
	}

	// append the data to the appender
	if err := appender(bts); err != nil {
		return err
	}

	// return the commit id
	return nil
}

func (k *AtomicKV) BeginApply(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.currentTx != nil {
		return ErrTxnActive
	}

	k.currentTx = k.db.BeginTransaction()

	return nil
}

func (k *AtomicKV) Apply(ctx context.Context, changes []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.currentTx == nil {
		return ErrTxnNotActive
	}

	// deserialize the changes
	values, err := types.DeserializeSlice[*keyValue](changes, func() *keyValue {
		return &keyValue{}
	})
	if err != nil {
		return err
	}

	// apply the changes
	for _, v := range values {
		var err error
		switch v.Operation {
		case kvOperationSet:
			err = k.currentTx.Set(v.Key, v.Value)
		case kvOperationDelete:
			err = k.currentTx.Delete(v.Key)
		default:
			err = fmt.Errorf("unknown operation")
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (k *AtomicKV) EndApply(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.currentTx == nil {
		return ErrTxnNotActive
	}

	err := k.currentTx.Commit()
	if err != nil {
		return err
	}

	k.currentTx = nil

	return nil
}

func (k *AtomicKV) Cancel(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.currentTx != nil {
		k.currentTx.Discard()
	}

	k.currentTx = nil
	k.inSession = false
	k.uncommittedData = make([]*keyValue, 0)
	return nil
}

func (k *AtomicKV) ID(ctx context.Context) ([]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.inSession {
		return nil, ErrSessionNotActive
	}

	hash := sha256.New()
	for _, v := range k.uncommittedData {
		bts, err := v.MarshalBinary()
		if err != nil {
			return nil, err
		}

		_, err = hash.Write(bts)
		if err != nil {
			return nil, err
		}
	}

	return hash.Sum(nil), nil
}

// flushUncommittedData returns the uncommitted data and clears the uncommitted data
func (k *AtomicKV) flushUncommittedData() ([]byte, error) {
	bts, err := types.SerializeSlice(k.uncommittedData)
	if err != nil {
		return nil, err
	}

	k.uncommittedData = make([]*keyValue, 0)
	return bts, nil
}
