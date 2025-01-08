package types

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	CodeInvalidSchema         TxCode = 100 // TODO: remove, as this is not applicable to the engine
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

// txResultsVer is the results structure or serialization version known presently
const txResultsVer uint16 = 0 // v0 has events with no data a future v1 will change how events are decoded

func (tr TxResult) MarshalBinary() ([]byte, error) {
	data := make([]byte, 2+4+4, 2+4+4+2+2) // put 10 bytes, append the rest

	// version
	binary.BigEndian.PutUint16(data, txResultsVer)

	// Encode code as 4 bytes
	binary.BigEndian.PutUint32(data[2:], tr.Code)

	// Encode log length as 4 bytes for length followed by log string
	binary.BigEndian.PutUint32(data[6:], uint32(len(tr.Log)))
	data = append(data, []byte(tr.Log)...)

	// Events
	numEvents := len(tr.Events)
	if numEvents > math.MaxUint16 {
		return nil, errors.New("too many events")
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
	if len(data) < 10 { // Minimum length: 2 bytes version + 4 bytes code + 4 bytes log length
		return errors.New("insufficient data")
	}

	var offset int

	version := binary.BigEndian.Uint16(data)
	if version != txResultsVer {
		return fmt.Errorf("unsupported version %d", version)
	}
	offset += 2

	// Decode code from first 4 bytes
	tr.Code = binary.BigEndian.Uint32(data[offset:])
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

// QueryResult is the result of a SQL query or action.
type QueryResult struct {
	ColumnNames []string    `json:"column_names"`
	ColumnTypes []*DataType `json:"column_types"`
	Values      [][]any     `json:"values"`
}

// ExportToStringMap converts the QueryResult to a slice of maps.
func (qr *QueryResult) ExportToStringMap() []map[string]string {
	var res []map[string]string
	for _, row := range qr.Values {
		m := make(map[string]string)
		for i, val := range row {
			m[qr.ColumnNames[i]] = fmt.Sprintf("%v", val)
		}
		res = append(res, m)
	}
	return res
}

// CallResult is the result of a procedure call.
type CallResult struct {
	QueryResult *QueryResult `json:"query_result"`
	Logs        []string     `json:"logs"`
}
