package numbers_test

import (
	"encoding/binary"
	"kwil/x/utils/numbers"
	"testing"
)

func Test_Bytes(t *testing.T) {
	// int64
	bts := numbers.Int64ToBytes(1)
	if len(bts) != 8 {
		t.Error("Expected 8 bytes")
	}
	if binary.BigEndian.Uint64(bts) != 1 {
		t.Error("Expected 1")
	}
	if numbers.BytesToInt64(bts) != 1 {
		t.Error("Expected 1")
	}

	// int32
	bts = numbers.Int32ToBytes(1)
	if len(bts) != 4 {
		t.Error("Expected 4 bytes")
	}
	if binary.BigEndian.Uint32(bts) != 1 {
		t.Error("Expected 1")
	}
	if numbers.BytesToInt32(bts) != 1 {
		t.Error("Expected 1")
	}

	// int16
	bts = numbers.Int16ToBytes(1)
	if len(bts) != 2 {
		t.Error("Expected 2 bytes")
	}
	if binary.BigEndian.Uint16(bts) != 1 {
		t.Error("Expected 1")
	}
	if numbers.BytesToInt16(bts) != 1 {
		t.Error("Expected 1")
	}

	// uint64
	bts = numbers.Uint64ToBytes(1)
	if len(bts) != 8 {
		t.Error("Expected 8 bytes")
	}
	if binary.BigEndian.Uint64(bts) != 1 {
		t.Error("Expected 1")
	}
	if numbers.BytesToUint64(bts) != 1 {
		t.Error("Expected 1")
	}

	// uint32
	bts = numbers.Uint32ToBytes(1)
	if len(bts) != 4 {
		t.Error("Expected 4 bytes")
	}
	if binary.BigEndian.Uint32(bts) != 1 {
		t.Error("Expected 1")
	}
	if numbers.BytesToUint32(bts) != 1 {
		t.Error("Expected 1")
	}

	// uint16
	bts = numbers.Uint16ToBytes(1)
	if len(bts) != 2 {
		t.Error("Expected 2 bytes")
	}
	if binary.BigEndian.Uint16(bts) != 1 {
		t.Error("Expected 1")
	}
	if numbers.BytesToUint16(bts) != 1 {
		t.Error("Expected 1")
	}
}
