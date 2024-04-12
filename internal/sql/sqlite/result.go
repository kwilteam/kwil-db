package sqlite

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"

	"github.com/kwilteam/go-sqlite"
)

type Result struct {
	// stmt is the connection to the actual sqlite statement.
	stmt *sqlite.Stmt

	// err is any error that occurred during iteration.
	err error

	// closed indicates whether the result set is closed.
	closed bool

	// complete indicates whether the result set has been completely read.
	complete bool

	// columnNames is the list of column names.
	columnNames []string

	// columnTypes is the list of column types.
	columnTypes []sqlite.ColumnType

	// firstIteration indicates whether this is the first iteration.
	// sqlite cannot return column types until the first iteration.
	firstIteration bool

	// closeFn is the function to closeFn the result set.
	// this is used to signal that the connection can be used again,
	// and is also used to close statements for read-only connections.
	closeFn func()

	// conn is the connection that this result set is associated with.
	conn *Connection
}

// Next steps to the next row.
func (r *Result) Next() (rowReturned bool) {
	if r.isClosed() {
		r.err = ErrClosed
		return false
	}

	if r.complete {
		return false
	}

	rowReturned, err := r.stmt.Step()
	if err != nil {
		if errors.Is(err, sqlite.ResultInterrupt.ToError()) {
			r.err = ErrInterrupted
		} else {
			r.err = err
		}
		return false
	}

	if r.firstIteration {
		r.firstIteration = false
		r.columnTypes = determineColumnTypes(r.stmt)
	}

	if !rowReturned {
		r.complete = true
	}

	return rowReturned
}

// Columns returns the column names of the current row.
func (r *Result) Columns() []string {

	return r.columnNames
}

// Values returns the values of the current row.
func (r *Result) Values() ([]any, error) {

	values := make([]any, len(r.columnNames))
	for i := range r.columnTypes {
		colType := r.stmt.ColumnType(i)
		switch colType {
		case sqlite.TypeInteger:
			values[i] = r.stmt.ColumnInt64(i)
		case sqlite.TypeFloat:
			float := r.stmt.ColumnFloat(i)
			if float == math.Trunc(float) {
				values[i] = int64(float)
			} else {
				return nil, ErrFloatDetected
			}
		case sqlite.TypeText:
			values[i] = r.stmt.ColumnText(i)
		case sqlite.TypeBlob:
			rdr := r.stmt.ColumnReader(i)
			bts, err := io.ReadAll(rdr)
			if err != nil {
				return nil, fmt.Errorf("kwildb getAny error: error reading blob: %w", err)
			}

			values[i] = bts
		case sqlite.TypeNull:
			values[i] = nil
		default:
			panic("kwildb get any error: unknown type")
		}
	}

	return values, nil
}

// Err gets any error that occurred during iteration.
func (r *Result) Err() error {
	return r.err
}

// Close closes the result set.
func (r *Result) Close() error {
	return r.close()
}

// close closes the result set.
// it does not acquire the result mutex.
func (r *Result) close() error {
	if r.isClosed() {
		return ErrClosed
	}

	r.closed = true
	r.closeFn()
	return nil
}

// Finish finishes any remaining execution and closes the result set.
func (r *Result) Finish() error {

	if r.isClosed() {
		return ErrClosed
	}

	// iterate through the result set to finish any remaining execution.
	for r.Next() {
	}
	if r.Err() != nil {
		return errors.Join(r.err, r.close())
	}

	return r.close()
}

func (r *Result) isClosed() bool {
	if r == nil {
		return true
	}

	if r.closed {
		return true
	}

	if r.conn.isClosed() {
		return true
	}

	return false
}

// setAny sets the given value to the parameter name.
// if the parameter does not exist, it will return nil.
func setAny(stmt *sqlite.Stmt, param string, val any) error {
	index := stmt.FindBindName("", param)
	if index <= 0 {
		return nil
	}

	ref := reflect.ValueOf(val)
	if !ref.IsValid() {
		stmt.SetNull(param)
		return nil
	}
	switch ref.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		stmt.BindInt64(index, ref.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		stmt.BindInt64(index, int64(ref.Uint()))
	case reflect.Float32, reflect.Float64:
		stmt.BindFloat(index, ref.Float())
	case reflect.String:
		stmt.BindText(index, ref.String())
	case reflect.Bool:
		stmt.BindBool(index, ref.Bool())
	case reflect.Array, reflect.Slice:
		stmt.BindBytes(index, ref.Bytes())
	default:
		return fmt.Errorf("kwildb set any error: unsupported type: %s", ref.Kind())
	}

	return nil
}

// setMany sets the given values to parameters based on their name.
// it will find the correct position of the parameter and bind the value to it.
func setMany(stmt *sqlite.Stmt, vals map[string]any) error {
	for param, val := range vals {
		if err := setAny(stmt, param, val); err != nil {
			return err
		}
	}
	return nil
}

// determineColumnNames determines the column names of the statement.
func determineColumnNames(stmt *sqlite.Stmt) []string {
	columnNames := make([]string, stmt.ColumnCount())
	for i := 0; i < len(columnNames); i++ {
		columnNames[i] = stmt.ColumnName(i)
	}
	return columnNames
}

// determineColumnTypes determines the column types of the statement.
func determineColumnTypes(stmt *sqlite.Stmt) []sqlite.ColumnType {
	columnTypes := make([]sqlite.ColumnType, stmt.ColumnCount())

	for i := 0; i < len(columnTypes); i++ {
		columnTypes[i] = stmt.ColumnType(i)
	}

	return columnTypes
}
