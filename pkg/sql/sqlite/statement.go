package sqlite

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/kwilteam/go-sqlite"
)

type Statement struct {
	busy        bool
	conn        *Connection
	stmt        *sqlite.Stmt
	columnNames []string
	columnTypes []DataType
}

func newStatement(conn *Connection, stmt *sqlite.Stmt) *Statement {
	s := &Statement{
		busy: false,
		conn: conn,
		stmt: stmt,
	}

	s.columnNames = determineColumnNames(s.stmt)

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
	// Args is a list of arguments to be passed to the query
	Args []interface{}

	// NamedArgs is a map of named arguments to be passed to the query
	NamedArgs map[string]interface{}
}

type ExecOption func(*ExecOpts)

// WithArgs specifies the args to use
func WithArgs(args ...interface{}) ExecOption {
	return func(opts *ExecOpts) {
		opts.Args = args
	}
}

// WithNamedArgs specifies the named args to use
func WithNamedArgs(namedArgs map[string]interface{}) ExecOption {
	return func(opts *ExecOpts) {
		opts.NamedArgs = namedArgs
	}
}

// Execute executes the statement.
// It takes an optional ExecOpts struct that can be used to set the arguments for the statement by parameter name,
// or by numeric index.  It also allows for a ResultFunc to be set that is called as the cursor steps through
// the result set.
// If both Args and NamedArgs are set, the NamedArgs will be used.
// Both NamedArgs and Args will override values set before the call to Execute.
// It also allows for a ResultSet to be set that will be populated with the results of the query.
func (s *Statement) Start(ctx context.Context, opts ...ExecOption) (*Results, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	return s.execute(ctx, opts...)
}

// internal execute function that does not lock the connection.
func (s *Statement) execute(ctx context.Context, options ...ExecOption) (*Results, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("connection has been closed")
	}

	if s.busy {
		return nil, fmt.Errorf("statement is busy")
	}

	s.busy = true

	opts := &ExecOpts{}

	for _, opt := range options {
		opt(opts)
	}

	err := s.bindParameters(opts)
	if err != nil {
		return nil, fmt.Errorf("error binding parameters: %w", err)
	}

	// Return results object to allow user to iterate manually
	return &Results{
		ctx:            ctx,
		statement:      s,
		firstIteration: true,
		options:        opts,
		closed:         false,
		complete:       false,
		closers:        make([]func() error, 0),
	}, nil
}

type Results struct {
	ctx            context.Context
	statement      *Statement
	firstIteration bool
	options        *ExecOpts
	closed         bool

	// complete tracks if there are more steps that can be taken
	complete bool

	// can be given extra closers for transient statements
	closers []func() error
}

func (r *Results) addCloser(closer func() error) {
	r.closers = append(r.closers, closer)
}

func (r *Results) Next() (rowReturned bool, err error) {
	if r.ctx != nil {
		select {
		case <-r.ctx.Done():
			return false, r.ctx.Err()
		default:
		}
	}

	if r.closed {
		return false, fmt.Errorf("results have been closed")
	}

	if r.complete {
		return false, nil
	}

	innerRowReturned, err := r.statement.step()
	if err != nil {
		return false, err
	}

	if !innerRowReturned {
		r.complete = true
		return false, nil
	}

	if r.firstIteration {
		r.statement.columnTypes = determineColumnTypes(r.statement.stmt)
		r.firstIteration = false
	}

	return true, nil
}

func (r *Results) GetRecord() map[string]any {
	return getRecord(r.statement.stmt, r.statement.columnNames, r.statement.columnTypes)
}

func getRecord(stmt *sqlite.Stmt, columnNames []string, columnTypes []DataType) map[string]any {
	row := make(map[string]any)
	for i, col := range columnNames {
		row[col] = getAny(stmt, i, columnTypes[i])
	}
	return row
}

// ExportRecords will export all records from the result set into a slice of maps.
// It closes the result set when it is done.
func (r *Results) ExportRecords() ([]map[string]any, error) {
	records := make([]map[string]any, 0)

	for {
		rowReturned, err := r.Next()
		if err != nil {
			return nil, err
		}

		if !rowReturned {
			break
		}

		records = append(records, r.GetRecord())
	}

	return records, r.Close()
}

// Finish finishes the results and closes the statement.
func (r *Results) Finish() (err error) {
	for {
		rowReturned, err := r.Next()
		if err != nil {
			return errors.Join(err, r.Close())
		}

		if !rowReturned {
			break
		}
	}

	return r.Close()
}

func (r *Results) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	r.statement.busy = false

	var errs []error
	err := r.statement.Clear()
	if err != nil {
		errs = append(errs, err)
	}

	for _, closer := range r.closers { // we need to run closers after clearing, since
		// closers are often used to close transient statements
		err := closer()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Reset resets the cursor to the beginning of the result set.
func (r *Results) Reset() error {
	return r.statement.Clear()
}

// bindParameters binds the parameters to the statement, whether they are named or not.
// it will also properly set the default values for global parameters.
// if there are conflicting values, the named parameters will override the positional parameters.
func (s *Statement) bindParameters(opts *ExecOpts) error {
	if opts.NamedArgs == nil {
		opts.NamedArgs = make(map[string]interface{})
	}

	err := bindMany(s.stmt, opts.Args)
	if err != nil {
		return fmt.Errorf("error binding args: %w", err)
	}

	// binding named after binding positional will override any positional values
	err = setMany(s.stmt, opts.NamedArgs)
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
func (s *Statement) GetBytes(param string) []byte {
	length := s.stmt.GetLen(param)

	buf := make([]byte, length)
	s.stmt.GetBytes(param, buf)

	return buf
}

// ReadBlob reads the blob value of the given parameter into the given byte slice.
func (s *Statement) ReadBlob(param string, buf []byte) {
	s.stmt.GetBytes(param, buf)
}

// getAny gets the value of the given parameter as an any.
func getAny(stmt *sqlite.Stmt, position int, typ DataType) any {
	switch typ {
	case DataTypeInteger:
		return stmt.ColumnInt64(position)
	case DataTypeFloat:
		return stmt.ColumnFloat(position)
	case DataTypeText:
		return stmt.ColumnText(position)
	case DataTypeBlob:
		rdr := stmt.ColumnReader(position)
		bts, err := io.ReadAll(rdr)
		if err != nil {
			panic(fmt.Errorf("kwildb get any error: error reading blob: %w", err))
		}
		return bts
	case DataTypeNull:
		return nil
	default:
		panic("kwildb get any error: unknown type")
	}
}

// bindAny binds the given value to the parameter index.
func bindAny(stmt *sqlite.Stmt, position int, val any) error {
	ref := reflect.ValueOf(val)
	if !ref.IsValid() {
		stmt.BindNull(position)
		return nil
	}
	switch ref.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		stmt.BindInt64(position, ref.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		stmt.BindInt64(position, int64(ref.Uint()))
	case reflect.Float32, reflect.Float64:
		stmt.BindFloat(position, ref.Float())
	case reflect.String:
		stmt.BindText(position, ref.String())
	case reflect.Bool:
		stmt.BindBool(position, ref.Bool())
	default:
		return fmt.Errorf("kwildb bind any error: unsupported type: %s", ref.Kind())
	}

	return nil
}

// bindMany binds the given values to parameters based on their index.
func bindMany(stmt *sqlite.Stmt, vals []any) error {
	for i, val := range vals {
		if err := bindAny(stmt, i+1, val); err != nil {
			return err
		}
	}
	return nil
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
	for i := 0; i < stmt.ColumnCount(); i++ {
		columnNames[i] = stmt.ColumnName(i)
	}
	return columnNames
}

// determineColumnTypes determines the column types of the statement.
func determineColumnTypes(stmt *sqlite.Stmt) []DataType {
	columnTypes := make([]DataType, stmt.ColumnCount())

	for i := 0; i < stmt.ColumnCount(); i++ {
		columnTypes[i] = convertColumnType(stmt.ColumnType(i))
	}

	return columnTypes
}

// Clear resets the statement and clears all bound parameters.
func (s *Statement) Clear() error {
	err := s.Reset()
	if err != nil {
		return err
	}

	return s.stmt.ClearBindings()
}

// Finalize finalizes the statement.
func (s *Statement) Finalize() error {
	return s.stmt.Finalize()
}

// Reset resets the statement.
func (s *Statement) Reset() error {
	return s.stmt.Reset()
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
