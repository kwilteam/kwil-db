package bytes

import "encoding/binary"

func Int16ToBytes(i int8) []byte {
	bts := make([]byte, 2)
	binary.BigEndian.PutUint16(bts, uint16(i))
	return bts
}

func Int32ToBytes(i int32) []byte {
	bts := make([]byte, 4)
	binary.BigEndian.PutUint32(bts, uint32(i))
	return bts
}

func Int64ToBytes(i int64) []byte {
	bts := make([]byte, 8)
	binary.BigEndian.PutUint64(bts, uint64(i))
	return bts
}

func Uint16ToBytes(i uint16) []byte {
	bts := make([]byte, 2)
	binary.BigEndian.PutUint16(bts, i)
	return bts
}

func Uint32ToBytes(i uint32) []byte {
	bts := make([]byte, 4)
	binary.BigEndian.PutUint32(bts, i)
	return bts
}

func Uint64ToBytes(i uint64) []byte {
	bts := make([]byte, 8)
	binary.BigEndian.PutUint64(bts, i)
	return bts
}

func BytesToInt16(bts []byte) int16 {
	return int16(binary.BigEndian.Uint16(bts))
}

func BytesToInt32(bts []byte) int32 {
	return int32(binary.BigEndian.Uint32(bts))
}

func BytesToInt64(bts []byte) int64 {
	return int64(binary.BigEndian.Uint64(bts))
}

func BytesToUint16(bts []byte) uint16 {
	return binary.BigEndian.Uint16(bts)
}

func BytesToUint32(bts []byte) uint32 {
	return binary.BigEndian.Uint32(bts)
}

func BytesToUint64(bts []byte) uint64 {
	return binary.BigEndian.Uint64(bts)
}
