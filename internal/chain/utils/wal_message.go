package utils

import "github.com/kwilteam/kwil-db/internal/common/utils"

type walMessage struct {
	data *[]byte
}

// , elems ...Type
func (m *walMessage) append(b ...byte) *walMessage {
	*(m.data) = append(*(m.data), b...)
	return m
}

func (m *walMessage) appendString(s string) *walMessage {
	*(m.data) = append(*(m.data), s...)
	return m
}

func (m *walMessage) appendLenWithString(s string) *walMessage {
	*(m.data) = append(*(m.data), utils.Uint16ToBytes(uint16(len(s)))...)
	return m
}

func (m *walMessage) appendUint64(n uint64) *walMessage {
	*(m.data) = append(*(m.data), utils.Uint64ToBytes(n)...)
	return m
}

func newWalMessage(msgType uint16) *walMessage {
	return &walMessage{newLogMsgPrefix(0, msgType)}
}

func newLogMsgPrefix(mByte uint8, msgType uint16) *[]byte {
	var m []byte
	m = append(m, mByte)
	m = append(m, utils.Uint16ToBytes(msgType)...)
	return &m
}
