package types

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type ConsensusReset struct {
	ToHeight int64
}

func (cr ConsensusReset) String() string {
	return fmt.Sprintf("ConsensusReset{Height: %d}", cr.ToHeight)
}

func (cr ConsensusReset) Bytes() []byte {
	return binary.LittleEndian.AppendUint64(nil, uint64(cr.ToHeight))
}

func (cr ConsensusReset) MarshalBinary() ([]byte, error) {
	return cr.Bytes(), nil
}

func (cr *ConsensusReset) UnmarshalBinary(data []byte) error {
	if len(data) != 8 {
		return errors.New("invalid ConsensusReset data")
	}
	cr.ToHeight = int64(binary.LittleEndian.Uint64(data))
	return nil
}

type AckRes struct {
	Height  int64
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
		return fmt.Sprintf("%s: height: %d, block %v, appHash %v", ar.ack(), ar.Height, ar.BlkHash, ar.AppHash)
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
	buf := make([]byte, 1+2*HashLen+8)
	buf[0] = 1
	binary.LittleEndian.PutUint64(buf[1:], uint64(ar.Height))
	copy(buf[1+8:], ar.BlkHash[:])
	copy(buf[1+8+HashLen:], ar.AppHash[:])
	return buf, nil
}

func (ar *AckRes) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("insufficient data")
	}
	ar.ACK = data[0] == 1
	if !ar.ACK {
		if len(data) > 1 {
			return errors.New("too much data for nACK")
		}
		ar.BlkHash = Hash{}
		ar.AppHash = nil
		return nil
	}
	data = data[1:]
	if len(data) < 2*HashLen+8 {
		return errors.New("insufficient data for ACK")
	}
	ar.Height = int64(binary.LittleEndian.Uint64(data[:8]))
	ar.AppHash = new(Hash)
	copy(ar.BlkHash[:], data[8:8+HashLen])
	copy(ar.AppHash[:], data[8+HashLen:])
	return nil
}
