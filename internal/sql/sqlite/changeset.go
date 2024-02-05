package sqlite

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/kwil-db/internal/utils/order"
)

// Changeset represents a set of changes to a database.
// For INSERTs, all inserted columns will be included.
// For UPDATEs, only the updated columns will be included.
// For DELETEs, no columns will be included.
type Changeset struct {
	// maps the table name to the table changeset
	Tables map[string]*TableChangeset
}

// TableChangeset represents a set of changes to a table.
type TableChangeset struct {
	// ColumnNames are the column names for the table.
	ColumnNames []string
	// maps the hex hash of the ordered primary keys to the record change
	Records map[string]*RecordChange
}

// RecordChange represents a change to a record.
type RecordChange struct {
	ChangeType RecordChangeType
	// maps the column name to the value
	Values []*Value
}

// RecordChangeType represents the type of change to a record.
type RecordChangeType uint8

const (
	RecordChangeTypeCreate RecordChangeType = iota
	RecordChangeTypeUpdate
	RecordChangeTypeDelete
)

// Value represents a value in a record.
type Value struct {
	DataType DataType
	Value    any
}

func (v *Value) Bytes() ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("kwildb: value is nil")
	}
	if v.DataType == DataTypeNull {
		return nil, nil
	}
	return v.DataType.ToBytes(v.Value)
}

// createChangeset creates a changeset from the given iterator.
func (s *Session) createChangeset(ctx context.Context, iter *sqlite.ChangesetIterator) (*Changeset, error) {
	tables := make(map[string]*TableChangeset)

	for {
		rowReturned, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if !rowReturned {
			break
		}

		operation, err := iter.Operation()
		if err != nil {
			return nil, err
		}

		primaryKeys, err := getPrimaryKeyValues(iter, operation)
		if err != nil {
			return nil, err
		}

		// tblColumns are the column names for the table
		tblColumns, err := s.conn.getColumnNames(ctx, operation.TableName)
		if err != nil {
			return nil, err
		}

		tableChangeset, ok := tables[operation.TableName]
		if !ok {
			tableChangeset = &TableChangeset{
				ColumnNames: tblColumns,
				Records:     make(map[string]*RecordChange),
			}
			tables[operation.TableName] = tableChangeset
		}

		recordChange := &RecordChange{
			Values: []*Value{},
		}

		var values []*Value
		switch operation.Type {
		default:
			panic(fmt.Sprintf("unknown operation type: %v", operation.Type))
		case sqlite.OpInsert:
			recordChange.ChangeType = RecordChangeTypeCreate
			values, err = extractValues(operation.NumColumns, iter.New)
		case sqlite.OpUpdate:
			recordChange.ChangeType = RecordChangeTypeUpdate
			values, err = extractValues(operation.NumColumns, iter.New)
		case sqlite.OpDelete:
			recordChange.ChangeType = RecordChangeTypeDelete
			// we don't need the values for a delete
		}
		if err != nil {
			return nil, err
		}

		recordChange.Values = values

		bts, err := ValueSet(primaryKeys).MarshalBinary()
		if err != nil {
			return nil, err
		}

		tableChangeset.Records[hex.EncodeToString(bts)] = recordChange
	}

	return &Changeset{
		Tables: tables,
	}, nil
}

// getPrimaryKeyValues returns the primary key values for the given iteration.
func getPrimaryKeyValues(iter *sqlite.ChangesetIterator, operation *sqlite.ChangesetOperation) ([]*Value, error) {
	primaryKeys := make([]*Value, 0)

	isPrimary, err := iter.PrimaryKey()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(isPrimary); i++ {
		var primaryVal *Value
		if isPrimary[i] {
			switch operation.Type {
			default:
				panic(fmt.Sprintf("unknown operation type: %v", operation.Type))
			case sqlite.OpInsert:
				value, err := iter.New(i)
				if err != nil {
					return nil, err
				}
				primaryVal, err = convertValue(&value)
				if err != nil {
					return nil, err
				}
			case sqlite.OpUpdate:
				value, err := iter.Old(i)
				if err != nil {
					return nil, err
				}
				primaryVal, err = convertValue(&value)
				if err != nil {
					return nil, err
				}
			case sqlite.OpDelete:
				value, err := iter.Old(i)
				if err != nil {
					return nil, err
				}
				primaryVal, err = convertValue(&value)
				if err != nil {
					return nil, err
				}
			}
			if err != nil {
				return nil, err
			}

			primaryKeys = append(primaryKeys, primaryVal)
		}
	}

	return primaryKeys, nil
}

// extractValues gets values from a function that returns a value for each column.
// it will not include null values.
func extractValues(numColumns int, fn func(int) (sqlite.Value, error)) ([]*Value, error) {
	values := make([]*Value, numColumns)
	for i := 0; i < numColumns; i++ {
		value, err := fn(i)
		if err != nil {
			return nil, err
		}

		val, err := convertValue(&value)
		if err != nil {
			return nil, err
		}

		values[i] = val
	}

	return values, nil
}

// convertValue converts a sqlite value to a Val.
func convertValue(v *sqlite.Value) (*Value, error) {
	var val *Value
	switch v.Type() {
	default:
		panic(fmt.Sprintf("unknown value type: %d", v.Type()))
	case sqlite.TypeInteger:
		val = &Value{
			DataType: DataTypeInt,
			Value:    v.Int64(),
		}
	case sqlite.TypeFloat:
		float := v.Float()
		if float == math.Trunc(float) {
			val = &Value{
				DataType: DataTypeInt,
				Value:    int64(float),
			}
		} else {
			return nil, ErrFloatDetected
		}
	case sqlite.TypeText:
		val = &Value{
			DataType: DataTypeText,
			Value:    v.Text(),
		}
	case sqlite.TypeBlob:
		b := make([]byte, len(v.Blob()))
		copy(b, v.Blob())
		val = &Value{
			DataType: DataTypeBlob,
			Value:    b,
		}
	case sqlite.TypeNull:
		val = &Value{
			DataType: DataTypeNull,
		}
	}

	return val, nil
}

// ValueSet represents a set of values.
type ValueSet []*Value

func (v ValueSet) MarshalBinary() ([]byte, error) {
	var b []byte
	for _, val := range v {
		bts, err := val.Bytes()
		if err != nil {
			return nil, err
		}
		length := uint32(len(bts))

		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, length)

		b = append(b, buf...)
		b = append(b, bts...)
	}

	return b, nil
}

func (v *ValueSet) UnmarshalBinary(b []byte) error {
	*v = make([]*Value, 0)

	for i := 0; i < len(b); {
		length := binary.LittleEndian.Uint32(b[i : i+4])
		i += 4

		dataType := DataType(b[i])
		i++

		val, err := dataType.FromBytes(b[i : i+int(length)])
		if err != nil {
			return err
		}

		*v = append(*v, &Value{
			DataType: dataType,
			Value:    val,
		})

		i += int(length)
	}

	return nil
}

// ID generates a unique ID for the changeset.
// It is deterministic and will always return the same ID for the same changeset.
// It will order the tables and records by name.
// It will keep each column sorted by index.
func (c *Changeset) ID() ([]byte, error) {
	hasher := sha256.New()

	for _, table := range order.OrderMap(c.Tables) {
		// table.Key = table name
		// table.value = table changeset
		for _, record := range order.OrderMap(table.Value.Records) {
			// record.Key = hex hash of the primary keys
			// record.value = record changeset

			for _, value := range record.Value.Values {
				bts, err := value.Bytes()
				if err != nil {
					return nil, err
				}

				hasher.Write(bts)
			}
		}
	}

	return hasher.Sum(nil), nil
}

type DataType uint8

const (
	DataTypeNull DataType = iota
	DataTypeInt
	DataTypeText
	DataTypeBlob
)

func (d DataType) FromBytes(b []byte) (any, error) {
	switch d {
	default:
		panic(fmt.Sprintf("unknown data type: %v", d))
	case DataTypeNull:
		return nil, nil
	case DataTypeInt:
		if len(b) != 8 {
			return nil, fmt.Errorf("expected 8 bytes for type INT, got %d", len(b))
		}
		return int64(binary.LittleEndian.Uint64(b)), nil
	case DataTypeText:
		return string(b), nil
	case DataTypeBlob:
		return b, nil
	}
}

func (d DataType) ToBytes(v any) ([]byte, error) {
	switch d {
	default:
		panic(fmt.Sprintf("unknown data type: %v", d))
	case DataTypeNull:
		return nil, nil
	case DataTypeInt:
		b := make([]byte, 8)
		switch t := v.(type) {
		default:
			return nil, fmt.Errorf("expected int64, int32, int16, int8, uint64, uint32, uint16, uint8, int, or uint for type INT, got %T", v)
		case int64:
			binary.LittleEndian.PutUint64(b, uint64(t))
		case int32:
			binary.LittleEndian.PutUint64(b, uint64(t))
		case int16:
			binary.LittleEndian.PutUint64(b, uint64(t))
		case int8:
			binary.LittleEndian.PutUint64(b, uint64(t))
		case uint64:
			binary.LittleEndian.PutUint64(b, t)
		case uint32:
			binary.LittleEndian.PutUint64(b, uint64(t))
		case uint16:
			binary.LittleEndian.PutUint64(b, uint64(t))
		case uint8:
			binary.LittleEndian.PutUint64(b, uint64(t))
		case int:
			binary.LittleEndian.PutUint64(b, uint64(t))
		case uint:
			binary.LittleEndian.PutUint64(b, uint64(t))
		}

		return b, nil
	case DataTypeText:
		strVal, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string for type TEXT, got %T", v)
		}

		return []byte(strVal), nil
	case DataTypeBlob:
		blobVal, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("expected []byte for type BLOB, got %T", v)
		}

		return blobVal, nil
	}
}
