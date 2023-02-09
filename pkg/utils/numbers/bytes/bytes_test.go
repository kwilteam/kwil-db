package bytes_test

import (
	"encoding/binary"
	"kwil/pkg/utils/numbers/bytes"
	"testing"
)

func Test_Bytes(t *testing.T) {
	// int64
	bts := bytes.Int64ToBytes(1)
	if len(bts) != 8 {
		t.Error("Expected 8 bytes")
	}
	if binary.BigEndian.Uint64(bts) != 1 {
		t.Error("Expected 1")
	}
	if bytes.BytesToInt64(bts) != 1 {
		t.Error("Expected 1")
	}

	// int32
	bts = bytes.Int32ToBytes(1)
	if len(bts) != 4 {
		t.Error("Expected 4 bytes")
	}
	if binary.BigEndian.Uint32(bts) != 1 {
		t.Error("Expected 1")
	}
	if bytes.BytesToInt32(bts) != 1 {
		t.Error("Expected 1")
	}

	// int16
	bts = bytes.Int16ToBytes(1)
	if len(bts) != 2 {
		t.Error("Expected 2 bytes")
	}
	if binary.BigEndian.Uint16(bts) != 1 {
		t.Error("Expected 1")
	}
	if bytes.BytesToInt16(bts) != 1 {
		t.Error("Expected 1")
	}

	// uint64
	bts = bytes.Uint64ToBytes(1)
	if len(bts) != 8 {
		t.Error("Expected 8 bytes")
	}
	if binary.BigEndian.Uint64(bts) != 1 {
		t.Error("Expected 1")
	}
	if bytes.BytesToUint64(bts) != 1 {
		t.Error("Expected 1")
	}

	// uint32
	bts = bytes.Uint32ToBytes(1)
	if len(bts) != 4 {
		t.Error("Expected 4 bytes")
	}
	if binary.BigEndian.Uint32(bts) != 1 {
		t.Error("Expected 1")
	}
	if bytes.BytesToUint32(bts) != 1 {
		t.Error("Expected 1")
	}

	// uint16
	bts = bytes.Uint16ToBytes(1)
	if len(bts) != 2 {
		t.Error("Expected 2 bytes")
	}
	if binary.BigEndian.Uint16(bts) != 1 {
		t.Error("Expected 1")
	}
	if bytes.BytesToUint16(bts) != 1 {
		t.Error("Expected 1")
	}
}
