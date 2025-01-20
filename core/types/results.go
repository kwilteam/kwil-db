package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
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

// CallResult is the result of a procedure call.
type CallResult struct {
	QueryResult *QueryResult `json:"query_result"`
	Logs        []string     `json:"logs"`
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

// Scan scans a value from the query result.
// It accepts a slice of pointers to values, and a function that will be called
// for each row in the result set.
// The passed values can be of type *string, *int64, *int, *bool, *[]byte, *UUID, *Decimal,
// *[]string, *[]int64, *[]int, *[]bool, *[]*int64, *[]*int, *[]*bool, *[]*UUID, *[]*Decimal,
// *[]UUID, *[]Decimal, *[][]byte, or *[]*[]byte.
func (q *QueryResult) Scan(fn func() error, vals ...any) error {
	for _, row := range q.Values {
		if err := ScanTo(row, vals...); err != nil {
			return err
		}
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ScanTo scans the src values into the dst values.
func ScanTo(src []any, dst ...any) error {
	if len(src) != len(dst) {
		return fmt.Errorf("expected %d columns, got %d", len(dst), len(src))
	}

	for j, col := range src {
		// if the column val is nil, we skip it.
		// If it is an array, we need to
		typeOf := reflect.TypeOf(col)
		if col == nil {
			continue
		} else if typeOf.Kind() == reflect.Slice && typeOf.Elem().Kind() != reflect.Uint8 {
			if err := convertArray(col, dst[j]); err != nil {
				return err
			}
			continue
		} else if typeOf.Kind() == reflect.Slice && typeOf.Elem().Kind() == reflect.Uint8 {
			if err := convertScalar(col, dst[j]); err != nil {
				return err
			}
			continue
		} else if typeOf.Kind() == reflect.Map {
			return fmt.Errorf("cannot scan value into map type: %T", dst[j])
		} else {
			if err := convertScalar(col, dst[j]); err != nil {
				return err
			}
		}
	}

	return nil
}

func convertArray(src any, dst any) error {
	arr, ok := src.([]any)
	if !ok {
		return fmt.Errorf("unexpected JSON array type: %T", src)
	}

	switch v := dst.(type) {
	case *[]string:
		return convArr(arr, v)
	case *[]*string:
		return convPtrArr(arr, v)
	case *[]int64:
		return convArr(arr, v)
	case *[]int:
		return convArr(arr, v)
	case *[]bool:
		return convArr(arr, v)
	case *[]*int64:
		return convPtrArr(arr, v)
	case *[]*int:
		return convPtrArr(arr, v)
	case *[]*bool:
		return convPtrArr(arr, v)
	case *[]*UUID:
		return convPtrArr(arr, v)
	case *[]UUID:
		return convArr(arr, v)
	case *[]*Decimal:
		return convPtrArr(arr, v)
	case *[]Decimal:
		return convArr(arr, v)
	case *[][]byte:
		return convArr(arr, v)
	case *[]*[]byte:
		return convPtrArr(arr, v)
	default:
		return fmt.Errorf("unexpected scan type: %T", dst)
	}
}

func convArr[T any](src []any, dst *[]T) error {
	dst2 := make([]T, len(src)) // we dont set the new slice to dst until we know we can convert all values
	for i, val := range src {
		if err := convertScalar(val, &dst2[i]); err != nil {
			return err
		}
	}
	*dst = dst2
	return nil
}

func convPtrArr[T any](src []any, dst *[]*T) error {
	dst2 := make([]*T, len(src)) // we dont set the new slice to dst until we know we can convert all values
	for i, val := range src {
		if val == nil {
			continue
		}

		s := new(T)

		err := convertScalar(val, s)
		if err != nil {
			return err
		}

		dst2[i] = s
	}
	*dst = dst2
	return nil
}

// convertScalar converts a scalar value to the specified type.
// It converts the source value to a string, then parses it into the specified type.
func convertScalar(src any, dst any) error {
	var null bool
	if src == nil {
		null = true
	}
	str, err := stringify(src)
	if err != nil {
		return err
	}
	switch v := dst.(type) {
	case *string:
		if null {
			return nil
		}
		*v = str
		return nil
	case *int64:
		if null {
			return nil
		}
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		*v = i
		return nil
	case *int:
		if null {
			return nil
		}
		i, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		*v = i
		return nil
	case *bool:
		if null {
			return nil
		}
		b, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		*v = b
		return nil
	case *[]byte:
		if null {
			return nil
		}
		*v = []byte(str)
		return nil
	case *UUID:
		if null {
			return nil
		}

		if len([]byte(str)) == 16 {
			*v = UUID([]byte(str))
			return nil
		}

		u, err := ParseUUID(str)
		if err != nil {
			return err
		}
		*v = *u
		return nil
	case *Decimal:
		if null {
			return nil
		}

		dec, err := ParseDecimal(str)
		if err != nil {
			return err
		}
		*v = *dec
		return nil
	default:
		return fmt.Errorf("unexpected scan type: %T", dst)
	}
}

// stringify converts a value as a string.
// It only expects values returned from JSON marshalling.
// It does NOT expect slices/arrays (except for []byte) or maps
func stringify(v any) (str string, err error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case []byte:
		return string(val), nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	case int:
		return strconv.Itoa(val), nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(val), nil
	case nil:
		return "", nil
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32), nil
	default:
		return "", fmt.Errorf("unexpected type: %T", v)
	}
}
