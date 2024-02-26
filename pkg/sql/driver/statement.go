package driver

import (
	"fmt"
	"reflect"

	"github.com/kwilteam/go-sqlite"
)

type Statement struct {
	stmt *sqlite.Stmt
}

func newStatement(stmt *sqlite.Stmt) *Statement {
	return &Statement{
		stmt: stmt,
	}
}

func (s *Statement) BindAny(position int, val any) error {
	ref := reflect.ValueOf(val)
	if !ref.IsValid() {
		s.stmt.BindNull(position)
		return nil
	}
	switch ref.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.stmt.BindInt64(position, ref.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s.stmt.BindInt64(position, int64(ref.Uint()))
	case reflect.Float32, reflect.Float64:
		return fmt.Errorf("floats are not supported")
	case reflect.String:
		s.stmt.BindText(position, ref.String())
	case reflect.Bool:
		s.stmt.BindBool(position, ref.Bool())
	default:
		return fmt.Errorf("sqlite bind any error: unsupported type: %s", ref.Kind())
	}

	return nil
}

func (s *Statement) BindMany(vals []any) error {
	for i, val := range vals {
		if err := s.BindAny(i+1, val); err != nil {
			return err
		}
	}
	return nil
}

func (s *Statement) BindInt(position int, val int) {
	s.stmt.BindInt64(position, int64(val))
}

func (s *Statement) BindText(position int, val string) {
	s.stmt.BindText(position, val)
}

func (s *Statement) BindBool(position int, val bool) {
	s.stmt.BindBool(position, val)
}

func (s *Statement) BindNull(position int) {
	s.stmt.BindNull(position)
}

func (s *Statement) BindBytes(position int, val []byte) {
	s.stmt.BindBytes(position, val)
}

func (s *Statement) SetAny(param string, val any) error {
	ref := reflect.ValueOf(val)
	if !ref.IsValid() {
		s.stmt.SetNull(param)
		return nil
	}
	switch ref.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.stmt.SetInt64(param, ref.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s.stmt.SetInt64(param, int64(ref.Uint()))
	case reflect.Float32, reflect.Float64:
		return fmt.Errorf("floats are not supported")
	case reflect.String:
		s.stmt.SetText(param, ref.String())
	case reflect.Bool:
		s.stmt.SetBool(param, ref.Bool())
	default:
		return fmt.Errorf("sqlite set any error: unsupported type: %s", ref.Kind())
	}

	return nil
}

func (s *Statement) SetMany(params map[string]any) error {
	for param, val := range params {
		if err := s.SetAny(param, val); err != nil {
			return err
		}
	}

	return nil
}

func (s *Statement) SetInt(param string, val int) {
	s.stmt.SetInt64(param, int64(val))
}

func (s *Statement) SetText(param string, val string) {
	s.stmt.SetText(param, val)
}

func (s *Statement) SetBool(param string, val bool) {
	s.stmt.SetBool(param, val)
}
func (s *Statement) SetNull(param string) {
	s.stmt.SetNull(param)
}

func (s *Statement) SetBytes(param string, val []byte) {
	s.stmt.SetBytes(param, val)
}

func (s *Statement) GetInt64(param string) int64 {
	return s.stmt.GetInt64(param)
}

func (s *Statement) GetText(param string) string {
	return s.stmt.GetText(param)
}

func (s *Statement) GetBool(param string) bool {
	return s.stmt.GetBool(param)
}

func (s *Statement) GetBytes(param string) (buf []byte, size int) {
	return buf, s.stmt.GetBytes(param, buf)
}

//func (s *Statement)

func (s *Statement) Clear() error {
	s.stmt.Reset()
	return s.stmt.ClearBindings()
}

// Step executes the statement and returns true if a row was returned.
// This is unexported to protect this packages consumer from violating
// a database's locking rules.
func (s *Statement) step() (rowReturned bool, err error) {
	return s.stmt.Step()
}

func (s *Statement) Columns() []string {
	colCount := s.stmt.ColumnCount()
	cols := make([]string, colCount)
	for i := 0; i < colCount; i++ {
		cols[i] = s.stmt.ColumnName(i)
	}

	return cols
}

type Record map[string]any

func (s *Statement) GetRecord() (Record, error) {
	colCount := s.stmt.ColumnCount()
	cols := make(Record)
	for i := 0; i < colCount; i++ {
		colName := s.stmt.ColumnName(i)
		colType := s.stmt.ColumnType(i)

		switch colType {
		case sqlite.TypeInteger:
			cols[colName] = s.stmt.GetInt64(colName)
		case sqlite.TypeFloat:
			cols[colName] = s.stmt.GetFloat(colName)
		case sqlite.TypeText:
			cols[colName] = s.stmt.GetText(colName)
		case sqlite.TypeBlob:
			cols[colName], _ = s.GetBytes(colName)
		case sqlite.TypeNull:
			cols[colName] = nil
		default:
			return nil, fmt.Errorf("unsupported column type: %s", colType)
		}
	}

	return cols, nil
}
