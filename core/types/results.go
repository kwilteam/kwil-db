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
func (q *QueryResult) Scan(vals []any, fn func() error) error {

	for _, row := range q.Values {
		if err := ScanTo(row, vals); err != nil {
			return err
		}
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// ScanTo scans the src values into the dst values.
func ScanTo(src []any, dst []any) error {
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
			if _, err := convertScalar(col, dst[j]); err != nil {
				return err
			}
			continue
		} else if typeOf.Kind() == reflect.Map {
			return fmt.Errorf("cannot scan value into map type: %T", dst[j])
		} else {
			if _, err := convertScalar(col, dst[j]); err != nil {
				return err
			}
		}
	}

	return nil
}

func convertArray(src any, dst any) error {
	var arr []any
	switch v := src.(type) {
	// most results will be []any, but in case it isn't we can reflect
	case []any:
		arr = v
	default:
		// otherwise, reflect
		val := reflect.ValueOf(src)
		if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
			return fmt.Errorf("unexpected type: %T", src)
		}
		arr = make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			arr[i] = val.Index(i).Interface()
		}
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
		if _, err := convertScalar(val, &dst2[i]); err != nil {
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

		wrote, err := convertScalar(val, s)
		if err != nil {
			return err
		}
		if !wrote {
			// if it didnt write a value, do not add the empty pointer to the slice
			continue
		}

		dst2[i] = s
	}
	*dst = dst2
	return nil
}

// convertScalar converts a scalar value to the specified type.
// It converts the source value to a string, then parses it into the specified type.
func convertScalar(src any, dst any) (wroteValue bool, err error) {
	var null bool
	str, null, err := stringify(src)
	if err != nil {
		return false, err
	}
	switch v := dst.(type) {
	case *string:
		if null {
			return false, nil
		}
		*v = str
		return true, nil
	case *int64:
		if null {
			return false, nil
		}
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return false, err
		}
		*v = i
		return true, nil
	case *int:
		if null {
			return false, nil
		}
		i, err := strconv.Atoi(str)
		if err != nil {
			return false, err
		}
		*v = i
		return true, nil
	case *bool:
		if null {
			return false, nil
		}
		b, err := strconv.ParseBool(str)
		if err != nil {
			return true, err
		}
		*v = b
		return true, nil
	case *[]byte:
		if null {
			return false, nil
		}
		*v = []byte(str)
		return true, nil
	case *UUID:
		if null {
			return false, nil
		}

		if len([]byte(str)) == 16 {
			*v = UUID([]byte(str))
			return true, nil
		}

		u, err := ParseUUID(str)
		if err != nil {
			return false, err
		}
		*v = *u
		return true, nil
	case *Decimal:
		if null {
			return false, nil
		}

		dec, err := ParseDecimal(str)
		if err != nil {
			return false, err
		}
		*v = *dec
		return true, nil
	default:
		return false, fmt.Errorf("unexpected scan type: %T", dst)
	}
}

// stringify converts a value as a string.
// It does NOT expect slices/arrays (except for []byte) or maps.
func stringify(v any) (str string, null bool, err error) {
	switch val := v.(type) {
	case string:
		return val, false, nil
	case []byte:
		return string(val), false, nil
	case int64:
		return strconv.FormatInt(val, 10), false, nil
	case int:
		return strconv.Itoa(val), false, nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), false, nil
	case bool:
		return strconv.FormatBool(val), false, nil
	case nil:
		return "", true, nil
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32), false, nil
	default:
		// if we hit here, we should see if it is a pointer, and if so, reflect and try again
		vOf := reflect.ValueOf(v)
		if vOf.Kind() == reflect.Ptr {
			elem := vOf.Elem()
			if elem.Kind() == reflect.Invalid {
				return "", true, nil
			}

			return stringify(elem.Interface())
		}
		return "", false, fmt.Errorf("unexpected type: %T", v)
	}
}
