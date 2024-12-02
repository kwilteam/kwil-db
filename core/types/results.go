package types

import (
	"encoding/binary"
	"errors"
	"math"
)

type TxCode uint16

const (
	CodeOk                  TxCode = 0
	CodeEncodingError       TxCode = 1
	CodeInvalidTxType       TxCode = 2
	CodeInvalidSignature    TxCode = 3
	CodeInvalidNonce        TxCode = 4
	CodeWrongChain          TxCode = 5
	CodeInsufficientBalance TxCode = 6
	CodeInsufficientFee     TxCode = 7
	CodeInvalidAmount       TxCode = 8
	CodeInvalidSender       TxCode = 9

	// engine-related error code
	CodeInvalidSchema         TxCode = 100
	CodeDatasetMissing        TxCode = 110
	CodeDatasetExists         TxCode = 120
	CodeInvalidResolutionType TxCode = 130

	CodeNetworkInMigration TxCode = 200

	CodeUnknownError TxCode = math.MaxUint16
)

var (
	// ErrTxNotFound is indicates when the a transaction was not found in the
	// nodes blocks or mempool.
	ErrTxNotFound          = errors.New("transaction not found")
	ErrWrongChain          = errors.New("wrong chain ID")
	ErrInvalidNonce        = errors.New("invalid nonce")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// TxResult is the result of a transaction execution on chain.
type TxResult struct {
	Code   uint32  `json:"code"`
	Gas    int64   `json:"gas"`
	Log    string  `json:"log"`
	Events []Event `json:"events"`
}

func (tr TxResult) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4+4, 4+4+2+2) // put 8 bytes, append the rest

	// Encode code as 4 bytes
	binary.BigEndian.PutUint32(data, tr.Code)

	// Encode log length as 4 bytes for length followed by log string
	binary.BigEndian.PutUint32(data[4:], uint32(len(tr.Log)))
	data = append(data, []byte(tr.Log)...)

	// Events
	numEvents := len(tr.Events)
	if numEvents > math.MaxUint16 {
		return nil, errors.New("to many events")
	}
	data = binary.BigEndian.AppendUint16(data, uint16(numEvents))
	for _, event := range tr.Events {
		evt, err := event.MarshalBinary()
		if err != nil {
			return nil, err
		}
		data = binary.BigEndian.AppendUint16(data, uint16(len(evt)))
		data = append(data, evt...)
	}

	return data, nil
}

func (tr *TxResult) UnmarshalBinary(data []byte) error {
	if len(data) < 8 { // Minimum length: 4 bytes code + 4 bytes log length
		return errors.New("insufficient data")
	}

	var offset int

	// Decode code from first 4 bytes
	tr.Code = binary.BigEndian.Uint32(data)
	offset += 4

	// Decode log length and string
	logLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if len(data) < offset+logLen {
		return errors.New("insufficient data for log")
	}
	tr.Log = string(data[offset : offset+logLen])
	offset += logLen

	// Decode events
	if len(data) < offset+2 {
		return errors.New("insufficient data for events length")
	}
	numEvents := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	tr.Events = make([]Event, numEvents)
	for i := range numEvents {
		if len(data) < offset+2 {
			return errors.New("insufficient data for event length")
		}
		eventLen := binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2

		if len(data) < offset+int(eventLen) {
			return errors.New("insufficient data for event")
		}
		if err := tr.Events[i].UnmarshalBinary(data[offset : offset+int(eventLen)]); err != nil {
			return err
		}
		offset += int(eventLen)
	}

	return nil
}

type Event struct{}

func (e Event) MarshalBinary() ([]byte, error) {
	return []byte{}, nil
}

func (e *Event) UnmarshalBinary(data []byte) error {
	return nil
}
