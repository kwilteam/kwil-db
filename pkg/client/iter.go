package client

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/cstockton/go-conv"
)

type Records struct {
	// index tracks the current row index for the iterator.
	index int

	// rows is the underlying sql.Rows object.
	records []*Record `json:"records"`
}

type Record map[string]any

func NewRecordFromMap(rec map[string]any) *Record {
	record := Record(rec)
	return &record
}

func NewRecords(records []*Record) *Records {
	return &Records{
		index:   -1,
		records: records,
	}
}

func NewRecordsFromMaps(recs []map[string]any) *Records {
	records := make([]*Record, len(recs))
	for i, rec := range recs {
		records[i] = NewRecordFromMap(rec)
	}

	return NewRecords(records)
}

func (r *Records) Next() bool {
	r.index++
	return r.index < len(r.records)
}

func (r *Records) Reset() {
	r.index = -1
}

func (r *Records) Record() *Record {
	return r.records[r.index]
}

func (r *Records) Scan(objects []interface{}) error {
	if !r.Next() {
		return errors.New("no more records")
	}

	if !isSliceOfPointers(objects) {
		return errors.New("objects must be a slice of pointers")
	}

	for i, record := range r.records {
		obj := objects[i]
		err := record.Scan(obj)
		if err != nil {
			return err
		}
	}

	return nil
}

func isSliceOfPointers(i interface{}) bool {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Slice {
		return false
	}

	return v.Type().Elem().Kind() == reflect.Ptr
}

func (r *Record) Scan(obj any) error {
	// check that obj is a pointer
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer")
	}

	val := reflect.ValueOf(obj)
	// check that obj is a struct
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Struct:
		return r.convertIntoStruct(val)
	case reflect.Slice:
		return r.convertIntoSlice(val)
	case reflect.Map:
		return r.convertIntoMap(val)
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.String, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool:
		return r.convertIntoScalarValue(val)
	default:
		return fmt.Errorf("record scan error: unsupported type: %s", reflect.TypeOf(obj).Kind().String())
	}
}

func (r *Record) convertIntoStruct(obj reflect.Value) error {
	objType := obj.Type()

	if obj.Kind() != reflect.Ptr || obj.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, got %s", objType)
	}

	structValue := obj.Elem()
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		fieldName := fieldType.Name
		mapValue, found := (*r)[fieldName]
		if !found {
			continue
		}

		fieldValue := reflect.ValueOf(mapValue)
		err := convertIntoScalar(field, fieldValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Record) convertIntoSlice(sliceValue reflect.Value) error {
	if sliceValue.IsNil() {
		sliceValue.Set(reflect.MakeSlice(sliceValue.Type(), 0, 0))
	}

	for _, value := range *r {
		sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(value)))
	}

	return nil
}

func (r *Record) convertIntoMap(mapValue reflect.Value) error {
	if mapValue.IsNil() {
		mapValue.Set(reflect.MakeMap(mapValue.Type()))
	}

	for key, value := range *r {
		mapValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}

	return nil
}

func (r *Record) convertIntoScalarValue(scalarValue reflect.Value) error {
	return convertIntoScalar(scalarValue, r.values())
}

func convertIntoScalar(scalarValue reflect.Value, val any) error {
	if !scalarValue.CanSet() {
		return errors.New("scalar value cannot be set")
	}

	switch scalarValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := conv.Int64(val)
		if err != nil {
			return err
		}

		scalarValue.SetInt(intVal)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := conv.Uint64(val)
		if err != nil {
			return err
		}

		scalarValue.SetUint(uintVal)
		return nil
	case reflect.Float32, reflect.Float64:
		floatVal, err := conv.Float64(val)
		if err != nil {
			return err
		}

		scalarValue.SetFloat(floatVal)
		return nil
	case reflect.String:
		stringVal, err := conv.String(val)
		if err != nil {
			return err
		}

		scalarValue.SetString(stringVal)
		return nil
	case reflect.Bool:
		boolVal, err := conv.Bool(val)
		if err != nil {
			return err
		}

		scalarValue.SetBool(boolVal)
		return nil
	default:
		return fmt.Errorf("scalar conversion error: unsupported type: %s", scalarValue.Kind().String())
	}
}

func (r Record) values() []any {
	var values []any
	for _, value := range r {
		values = append(values, value)
	}
	return values
}
