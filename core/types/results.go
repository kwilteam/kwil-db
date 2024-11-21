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

type TxResult struct {
	Code   uint16
	Gas    int64
	Log    string
	Events []Event
}

func (tr TxResult) MarshalBinary() ([]byte, error) {
	data := make([]byte, 2+4, 2+4+4) // put 6 bytes, append the rest

	// Encode code as 2 bytes
	binary.BigEndian.PutUint16(data, tr.Code)

	// Encode log length as 4 bytes for length followed by log string
	binary.BigEndian.PutUint32(data[2:], uint32(len(tr.Log)))
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
	if len(data) < 6 { // Minimum length: 2 bytes code + 4 bytes log length
		return errors.New("insufficient data")
	}

	// Decode code from first 2 bytes
	tr.Code = binary.BigEndian.Uint16(data[:2])

	// Decode log length and string
	logLen := int(binary.BigEndian.Uint32(data[2:]))
	if len(data) < 6+logLen {
		return errors.New("insufficient data for log")
	}
	tr.Log = string(data[6 : 6+logLen])

	// Move cursor past the log
	cursor := 6 + logLen

	// Decode events
	if len(data) < cursor+2 {
		return errors.New("insufficient data for events length")
	}
	numEvents := binary.BigEndian.Uint16(data[cursor : cursor+2])
	cursor += 2

	tr.Events = make([]Event, numEvents)
	for i := range numEvents {
		if len(data) < cursor+2 {
			return errors.New("insufficient data for event length")
		}
		eventLen := binary.BigEndian.Uint16(data[cursor : cursor+2])
		cursor += 2

		if len(data) < cursor+int(eventLen) {
			return errors.New("insufficient data for event")
		}
		if err := tr.Events[i].UnmarshalBinary(data[cursor : cursor+int(eventLen)]); err != nil {
			return err
		}
		cursor += int(eventLen)
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
