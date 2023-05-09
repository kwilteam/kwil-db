package sqlite

import (
	"fmt"
	"reflect"

	"github.com/kwilteam/go-sqlite"
)

type Statement struct {
	conn        *Connection
	stmt        *sqlite.Stmt
	columnNames []string
	columnTypes []DataType
}

func newStatement(conn *Connection, stmt *sqlite.Stmt) *Statement {
	s := &Statement{
		conn: conn,
		stmt: stmt,
	}

	s.determineColumnNames()

	return s
}

/*
// SetInt64 sets the value of the given parameter to the given int64 value.
func (s *Statement) SetInt(param string, val int64) {
	index := s.stmt.FindBindName("SetInt64", param)
	if index == 0 {
		return
	}

	s.stmt.BindInt64(index, val)
}

// SetText sets the value of the given parameter to the given string value.
func (s *Statement) SetText(param string, val string) {
	index := s.stmt.FindBindName("SetText", param)
	if index == 0 {
		return
	}

	s.stmt.BindText(index, val)
}

// SetBytes sets the value of the given parameter to the given byte slice value.
func (s *Statement) SetBytes(param string, val []byte) {
	index := s.stmt.FindBindName("SetBytes", param)
	if index == 0 {
		return
	}

	s.stmt.BindBytes(index, val)
}

// SetBool sets the value of the given parameter to the given bool value.
func (s *Statement) SetBool(param string, val bool) {
	index := s.stmt.FindBindName("SetBool", param)
	if index == 0 {
		return
	}

	s.stmt.BindBool(index, val)
}
*/

// step steps to the next row in the result set.
func (s *Statement) step() (rowReturned bool, err error) {
	return s.stmt.Step()
}

type ExecOpts struct {
	//ResultFunc is a function that is called for each row returned
	ResultFunc func(*Statement) error

	// Args is a list of arguments to be passed to the query
	Args []interface{}

	// NamedArgs is a map of named arguments to be passed to the query
	NamedArgs map[string]interface{}

	// ResultSet
	ResultSet *ResultSet
}

// addDefaults adds the named arguments to the ExecOpts if they are not already set.
func (e *ExecOpts) addDefaults(defaults map[string]any) {
	if e.NamedArgs == nil {
		e.NamedArgs = make(map[string]interface{})
	}

	for k, v := range defaults {
		if _, ok := e.NamedArgs[k]; ok {
			continue
		}

		e.NamedArgs[k] = v
	}
}

func (e *ExecOpts) ensureResultFunc() {
	if e.ResultFunc == nil {
		e.ResultFunc = func(*Statement) error {
			return nil
		}
	}
}

// Execute executes the statement.
// It takes an optional ExecOpts struct that can be used to set the arguments for the statement by parameter name,
// or by numeric index.  It also allows for a ResultFunc to be set that is called as the cursor steps through
// the result set.
// If both Args and NamedArgs are set, the NamedArgs will be used.
// Both NamedArgs and Args will override values set before the call to Execute.
// It also allows for a ResultSet to be set that will be populated with the results of the query.
func (s *Statement) Execute(opts *ExecOpts) error {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	return s.execute(opts)
}

// internal execute function that does not lock the connection.
func (s *Statement) execute(opts *ExecOpts) error {
	defer s.Clear()

	if s.conn == nil {
		return fmt.Errorf("connection has been closed")
	}

	if opts == nil {
		opts = &ExecOpts{}
	}

	opts.ensureResultFunc()

	err := s.bindParameters(opts)
	if err != nil {
		return fmt.Errorf("error binding parameters: %w", err)
	}

	useResultSet := false
	if opts.ResultSet != nil {
		useResultSet = true

		opts.ResultSet.index = -1
		opts.ResultSet.Columns = s.columnNames
	}

	firstIteration := true
	for {
		rowReturned, err := s.step()
		if err != nil {
			return err
		}

		if !rowReturned {
			break
		}

		if firstIteration {
			s.determineColumnTypes() // sqlite doesn't detect proper column types on the 0th iteration
			firstIteration = false
		}

		if useResultSet {
			opts.ResultSet.Rows = append(opts.ResultSet.Rows, s.getRow())
		}

		err = opts.ResultFunc(s)
		if err != nil {
			return err
		}
	}

	return nil
}

// bindParameters binds the parameters to the statement, whether they are named or not.
// it will also properly set the default values for global parameters.
// if there are conflicting values, the named parameters will override the positional parameters.
func (s *Statement) bindParameters(opts *ExecOpts) error {
	if opts.NamedArgs == nil {
		opts.NamedArgs = make(map[string]interface{})
	}

	opts.addDefaults(s.conn.globalVariableMap)

	err := s.bindMany(opts.Args)
	if err != nil {
		return fmt.Errorf("error binding args: %w", err)
	}

	// binding named after binding positional will override any positional values
	err = s.setMany(opts.NamedArgs)
	if err != nil {
		return fmt.Errorf("error setting named args: %w", err)
	}

	return nil
}

// GetText gets the string value of the given parameter.
func (s *Statement) GetText(param string) string {
	return s.stmt.GetText(param)
}

// GetInt64 gets the int value of the given parameter.
func (s *Statement) GetInt64(param string) int64 {
	return s.stmt.GetInt64(param)
}

// GetFloat64 gets the float64 value of the given parameter.
func (s *Statement) GetFloat64(param string) float64 {
	return s.stmt.GetFloat(param)
}

// GetBool gets the bool value of the given parameter.
func (s *Statement) GetBool(param string) bool {
	return s.stmt.GetBool(param)
}

// GetBytes gets the blob value of the given parameter and returns it as a byte slice.
func (s *Statement) GetBytes(param string) (buf []byte) {
	s.stmt.GetBytes(param, buf)
	return buf
}

// ReadBlob reads the blob value of the given parameter into the given byte slice.
func (s *Statement) ReadBlob(param string, buf []byte) {
	s.stmt.GetBytes(param, buf)
}

func (s *Statement) getRow() []any {
	row := make([]any, len(s.columnTypes))
	for i := range row {
		row[i] = s.getAny(i)
	}
	return row
}

func (s *Statement) getAny(position int) any {
	switch s.columnTypes[position] {
	case DataTypeInteger:
		return s.stmt.ColumnInt64(position)
	case DataTypeFloat:
		return s.stmt.ColumnFloat(position)
	case DataTypeText:
		return s.stmt.ColumnText(position)
	case DataTypeBlob:
		var buf []byte
		s.stmt.ColumnBytes(position, buf)
		return buf
	case DataTypeNull:
		return nil
	default:
		panic("kwildb get any error: unknown type")
	}
}

// bindAny binds the given value to the parameter index.
func (s *Statement) bindAny(position int, val any) error {
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
		s.stmt.BindFloat(position, ref.Float())
	case reflect.String:
		s.stmt.BindText(position, ref.String())
	case reflect.Bool:
		s.stmt.BindBool(position, ref.Bool())
	default:
		return fmt.Errorf("kwildb bind any error: unsupported type: %s", ref.Kind())
	}

	return nil
}

// bindMany binds the given values to parameters based on their index.
func (s *Statement) bindMany(vals []any) error {
	for i, val := range vals {
		if err := s.bindAny(i+1, val); err != nil {
			return err
		}
	}
	return nil
}

// setAny sets the given value to the parameter name.
func (s *Statement) setAny(param string, val any) error {
	index := s.stmt.FindBindName("kwil set any", param)
	if index <= 0 {
		return nil
	}

	ref := reflect.ValueOf(val)
	if !ref.IsValid() {
		s.stmt.SetNull(param)
		return nil
	}
	switch ref.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.stmt.BindInt64(index, ref.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s.stmt.BindInt64(index, int64(ref.Uint()))
	case reflect.Float32, reflect.Float64:
		s.stmt.BindFloat(index, ref.Float())
	case reflect.String:
		s.stmt.BindText(index, ref.String())
	case reflect.Bool:
		s.stmt.BindBool(index, ref.Bool())
	default:
		return fmt.Errorf("kwildb set any error: unsupported type: %s", ref.Kind())
	}

	return nil
}

// setMany sets the given values to parameters based on their name.
func (s *Statement) setMany(vals map[string]any) error {
	for param, val := range vals {
		if err := s.setAny(param, val); err != nil {
			return err
		}
	}
	return nil
}

// determineColumnNames determines the column names of the statement.
func (s *Statement) determineColumnNames() {
	if s.columnNames == nil {
		s.columnNames = make([]string, s.stmt.ColumnCount())
	}
	for i := 0; i < s.stmt.ColumnCount(); i++ {
		s.columnNames[i] = s.stmt.ColumnName(i)
	}
}

// determineColumnTypes determines the column types of the statement.
func (s *Statement) determineColumnTypes() {
	s.columnTypes = make([]DataType, s.stmt.ColumnCount())

	for i := 0; i < s.stmt.ColumnCount(); i++ {
		s.columnTypes[i] = convertColumnType(s.stmt.ColumnType(i))
	}
}

// Clear resets the statement and clears all bound parameters.
func (s *Statement) Clear() error {
	err := s.stmt.Reset()
	if err != nil {
		return err
	}

	return s.stmt.ClearBindings()
}

// Finalize finalizes the statement.
func (s *Statement) Finalize() error {
	return s.stmt.Finalize()
}

// convertColumnType converts a sqlite.ColumnType to a DataType.
func convertColumnType(typ1 sqlite.ColumnType) DataType {
	switch typ1 {
	case sqlite.TypeInteger:
		return DataTypeInteger
	case sqlite.TypeFloat:
		return DataTypeFloat
	case sqlite.TypeText:
		return DataTypeText
	case sqlite.TypeBlob:
		return DataTypeBlob
	case sqlite.TypeNull:
		return DataTypeText // this is not a bug, if the user typecasts a return value it will get reported as type null and not be read properly
		// if the value is actually null, it will just be read as a string
	default:
		return DataTypeText
	}
}
