package types

import "fmt"

type AckRes struct {
	ACK     bool
	BlkHash Hash
	AppHash Hash
}

func (ar AckRes) ack() string {
	if ar.ACK {
		return "ACK"
	}
	return "nACK"
}

func (ar AckRes) String() string {
	if ar.ACK {
		return fmt.Sprintf("%s: block %d, appHash %x", ar.ack(), ar.BlkHash, ar.AppHash)
	}
	return ar.ack()
}

func (ar AckRes) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 1+len(ar.BlkHash)+len(ar.AppHash))
	if ar.ACK {
		buf[0] = 1
	}
	copy(buf[1:], ar.BlkHash[:])
	copy(buf[1+len(ar.BlkHash)+4:], ar.AppHash[:])
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
		ar.AppHash = Hash{}
		return nil
	}
	data = data[1:]
	if len(data) < len(ar.BlkHash)+len(ar.AppHash) {
		return fmt.Errorf("insufficient data for ACK")
	}
	copy(ar.BlkHash[:], data[:len(ar.BlkHash)])
	copy(ar.AppHash[:], data[len(ar.BlkHash):])
	return nil
}
