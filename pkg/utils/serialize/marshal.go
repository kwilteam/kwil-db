package serialize

import (
	"encoding/binary"
)

func StringToBytes(s string) []byte {
	return []byte(s)
}

func BytesToString(b []byte) string {
	return string(b)
}

func Int32ToBytes(i int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(i))
	return b
}

func BytesToInt32(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(b))
}

func Int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}

func BytesToInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

func BoolToBytes(b bool) []byte {
	if b {
		return []byte{1}
	}
	return []byte{0}
}

func BytesToBool(b []byte) bool {
	return b[0] == 1
}
