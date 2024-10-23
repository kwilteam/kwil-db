package types

import (
	"errors"
	"fmt"
)

type AckRes struct {
	ACK     bool
	BlkHash Hash
	AppHash *Hash
}

func (ar AckRes) ack() string {
	if ar.ACK {
		return "ACK"
	}
	return "nACK"
}

func (ar AckRes) String() string {
	if ar.ACK {
		return fmt.Sprintf("%s: block %v, appHash %v", ar.ack(), ar.BlkHash, ar.AppHash)
	}
	return ar.ack()
}

func (ar AckRes) MarshalBinary() ([]byte, error) {
	if !ar.ACK {
		return []byte{0}, nil
	}
	if ar.AppHash == nil {
		return nil, errors.New("missing apphash in ACK")
	}
	buf := make([]byte, 1+2*HashLen)
	buf[0] = 1
	copy(buf[1:], ar.BlkHash[:])
	copy(buf[1+len(ar.BlkHash):], ar.AppHash[:])
	return buf, nil
}

func (ar *AckRes) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return fmt.Errorf("insufficient data")
	}
	ar.ACK = data[0] == 1
	if !ar.ACK {
		if len(data) > 1 {
			return fmt.Errorf("too much data for nACK")
		}
		ar.BlkHash = Hash{}
		ar.AppHash = nil
		return nil
	}
	data = data[1:]
	if len(data) < 2*HashLen {
		return fmt.Errorf("insufficient data for ACK")
	}
	ar.AppHash = new(Hash)
	copy(ar.BlkHash[:], data[:len(ar.BlkHash)])
	copy(ar.AppHash[:], data[len(ar.BlkHash):])
	return nil
}
