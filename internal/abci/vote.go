package abci

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	voteExtVer        = 0 // encoding of raw vote extension
	voteExtSegmentVer = 0 // encoding of a segment
)

type Serializer []byte

func (s Serializer) Append(data []byte) Serializer {
	s = binary.AppendUvarint(s, uint64(len(data)))
	return append(s, data...)
}

func (s Serializer) AppendUint16(val uint16) Serializer {
	return binary.BigEndian.AppendUint16(s, val)
}

func (s Serializer) AppendUint32(val uint32) Serializer {
	return binary.BigEndian.AppendUint32(s, val)
}

func (s Serializer) AppendUint64(val uint64) Serializer {
	return binary.BigEndian.AppendUint64(s, val)
}

// hint: math.Float64bits(f)

func (s Serializer) AppendBool(b bool) []byte {
	if b {
		return append(s, byte(1))
	}
	return append(s, byte(0))
}

func ExtractDataSegments(b []byte) ([][]byte, error) {
	var data [][]byte
	for {
		if len(b) == 0 {
			break
		}
		l, n := binary.Uvarint(b)
		if n <= 0 {
			return nil, errors.New("invalid varint")
		}

		b = b[n:]
		if len(b) < int(l) {
			return nil, fmt.Errorf("data too short for %d bytes", l)
		}
		if l == 0 {
			data = append(data, nil) // nil, not empty slice
			continue
		}
		data = append(data, b[:l])
		b = b[l:]
	}
	return data, nil
}

func DecodeVersionedData(b []byte) (uint8, [][]byte, error) {
	if len(b) == 0 {
		return 0, nil, fmt.Errorf("empty data")
	}
	ver, b := b[0], b[1:]
	segments, err := ExtractDataSegments(b)
	return ver, segments, err
}

var ExtIntCoder = binary.BigEndian

// In VerifyVoteExtension, we decode and check the vote extension from other
// validators, and dispatch to other modules depending on the type of the
// segments. For instance, if an vote extension decodes a token bridge event
// type, it will be passed to the BridgeEventsModule for processing.
//
// TODO: If the above is sensible, consider a system that allows to expand to
// other vote types easily. type VoteExtensionsMgr struct{}

// VoteExtensionSegment is a segment of a serialized vote extension payload, as
// received by VerifyVoteExtension or sent by ExtendVote.
type VoteExtensionSegment struct {
	Version uint16
	Type    uint32
	Data    []byte // decodes into a specific extension type
}

type VoteExtension []*VoteExtensionSegment

type VoteExtensionType = uint32

const (
	VoteExtensionTypeDeposit VoteExtensionType = 1 // special case of chain event for token bridge linked to Kwil chain gas
	VoteExtensionTypeTest    VoteExtensionType = 2
	// VoteExtensionTypeChainEvt VoteExtensionType = 3 // generalized chain events, recognized by ...?
)

// type VoteExtSegment interface {
// 	Type() uint32
// 	encoding.BinaryMarshaler
// 	encoding.BinaryUnmarshaler
// }

// type VoteExt []VoteExtensionSegment

func EncodeVoteExtension(vExt VoteExtension) []byte {
	// Encode in the convention used by DecodeVersionedData+ExtractDataSegments.
	ve := Serializer{voteExtVer}
	for _, segment := range vExt {
		ve = ve.Append(encodeVoteExtensionSegment(segment))
	}
	return ve
}

func encodeVoteExtensionSegment(vExt *VoteExtensionSegment) []byte {
	if vExt == nil {
		return nil
	}
	const hdrFieldLen = 2 + 4 // 2 bytes for the version, 4 bytes for the type
	b := make([]byte, hdrFieldLen, hdrFieldLen+len(vExt.Data))
	ExtIntCoder.PutUint16(b, vExt.Version)
	ExtIntCoder.PutUint32(b[2:], vExt.Type)
	return append(b, vExt.Data...)
}

// DecodeVoteExtension decodes a vote extension received by VerifyVoteExtension.
// The version and type should be used to decode the vote data and invoke the
// appropriate module.
func DecodeVoteExtension(ve []byte) (VoteExtension, error) {
	if len(ve) == 0 {
		return nil, nil
	}
	if len(ve) < 2 {
		return nil, fmt.Errorf("invalid vote extension")
	}

	ver, segments, err := DecodeVersionedData(ve)
	if err != nil {
		return nil, err
	}
	if ver != voteExtVer {
		return nil, fmt.Errorf("unsupported vote extension serialization version %d", ver)
	} // or switch to different format

	exts := make([]*VoteExtensionSegment, len(segments))
	for i := range segments {
		exts[i], err = decodeVoteExtensionSegment(segments[i])
		if err != nil {
			return nil, err
		}
	}
	return exts, nil
}

func decodeVoteExtensionSegment(ve []byte) (*VoteExtensionSegment, error) {
	if len(ve) == 0 {
		return nil, nil
	}
	if len(ve) < 2 {
		return nil, fmt.Errorf("invalid vote extension segment")
	}

	ver := ExtIntCoder.Uint16(ve)
	if ver != voteExtSegmentVer {
		return nil, fmt.Errorf("unsupported vote extension version: %d", ver)
	}

	ve = ve[2:]

	if len(ve) < 4+1 { // 4 bytes for the type, and at least 1 byte for the data
		return nil, fmt.Errorf("invalid vote extension segment")
	}

	vExt := &VoteExtensionSegment{
		Version: ver,
		Type:    ExtIntCoder.Uint32(ve),
		Data:    ve[4:],
	}

	return vExt, nil
}

// type ChainEventVoteExt struct {
// 	Type uint32
// 	ID   string
// 	Data []byte // serialization of the event data like {acct,amt} for a deposit?
// }
//   ^ in this case, how does abci (or a module) decode the acct and amt to credit?
//     would it use a chain package to decode it too?

type DepositVoteExt struct {
	EventID string
	Account string
	Amount  string // decode into big.Int with SetString(Amount, 10)
}

// func (d *DepositVoteExt) Type() uint32 {
// 	return VoteExtensionTypeDeposit
// }

func (d *DepositVoteExt) MarshalBinary() ([]byte, error) {
	return Serializer{0}. // v0
				Append([]byte(d.EventID)).
				Append([]byte(d.Account)).
				Append([]byte(d.Amount)), nil
}

// Bytes is a convenience since MarshalBinary does not error.
func (d *DepositVoteExt) Bytes() []byte {
	b, _ := d.MarshalBinary()
	return b
}

func (d *DepositVoteExt) EncodeSegment() *VoteExtensionSegment {
	return &VoteExtensionSegment{
		Version: voteExtSegmentVer,
		Type:    VoteExtensionTypeDeposit,
		Data:    d.Bytes(),
	}
}

// UnmarshalBinary deserializes from the Data field of a vote extension segment.
func (d *DepositVoteExt) UnmarshalBinary(b []byte) error {
	ver, fields, err := DecodeVersionedData(b)
	if err != nil {
		return err
	}
	if ver != 0 {
		return fmt.Errorf("unrecognized serialization version %d", ver)
	} // if serialization changes, switch on version to use different decoding

	const needFields = 3
	if len(fields) != needFields {
		return fmt.Errorf("expected %d fields, got %d", needFields, len(fields))
	}

	d.EventID = string(fields[0])
	d.Account = string(fields[1])
	d.Amount = string(fields[2])
	return nil
}

type TestVoteExt struct {
	Msg string
}

// func (d *TestVoteExt) Type() uint32 {
// 	return VoteExtensionTypeTest
// }

func (d *TestVoteExt) MarshalBinary() ([]byte, error) {
	return Serializer{0}.Append([]byte(d.Msg)), nil
}

func (d *TestVoteExt) UnmarshalBinary(b []byte) error {
	ver, fields, err := DecodeVersionedData(b)
	if err != nil {
		return err
	}
	if ver != 0 {
		return fmt.Errorf("unrecognized serialization version %d", ver)
	}
	const needFields = 1
	if len(fields) != needFields {
		return fmt.Errorf("expected %d fields, got %d", needFields, len(fields))
	}
	d.Msg = string(fields[0])
	return nil
}
