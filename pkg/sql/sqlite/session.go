package sqlite

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/kwilteam/go-sqlite"
)

// Session is a session for a database.
// It can be used to track changes and make a changeset.
type Session struct {
	ses *sqlite.Session
}

// CreateSession creates a new session.
// The sessions tracks all changes made to the database.
func (c *Connection) CreateSession() (*Session, error) {
	ses, err := c.conn.CreateSession("")
	if err != nil {
		return nil, err
	}

	// attaches all tables
	err = ses.Attach("")
	if err != nil {
		return nil, err
	}

	return &Session{
		ses: ses,
	}, nil
}

// Delete deletes the session and associated resources.
func (s *Session) Delete() {
	s.ses.Delete()
}

// GenerateChangeset generates a changeset for the session.
// Ensure that you close the changeset when you are done with it.
func (s *Session) GenerateChangeset() (*Changeset, error) {
	buf := new(bytes.Buffer)
	err := s.ses.WriteChangeset(buf)
	if err != nil {
		return nil, err
	}

	return NewChangset(buf)
}

var (
	ErrForeignKeyConflict = errors.New("foreign key conflict")
)

// NewChangset creates a new changeset from bytes.
func NewChangset(buf *bytes.Buffer) (*Changeset, error) {
	iter, err := sqlite.NewChangesetIterator(buf)
	if err != nil {
		return nil, err
	}

	return &Changeset{
		buf:  buf,
		iter: iter,
	}, nil
}

// Changeset is a changeset generated from a session.
type Changeset struct {
	buf  *bytes.Buffer
	iter *sqlite.ChangesetIterator
}

// Next returns true if there is another row in the changeset.
// If there is a foreign key conflict, it will return false and ErrForeignKeyConflict.
func (c *Changeset) Next() (rowReturned bool, err error) {
	rowReturned, err = c.iter.Next()
	if err != nil {
		return false, fmt.Errorf("Changeset.Next(): failed to get next row: %w", err)
	}

	return rowReturned, nil
}

// Close closes the changeset.
func (c *Changeset) Close() error {
	return c.iter.Close()
}

// Operation returns the operation of the current row.
func (c *Changeset) Operation() (*ChangesetOperation, error) {
	innerOperation, err := c.iter.Operation()
	if err != nil {
		return nil, err
	}

	opType, ok := innerOpTypeMap[innerOperation.Type]
	if !ok {
		return nil, fmt.Errorf("unknown operation type received from database engine: %v", innerOperation.Type)
	}

	return &ChangesetOperation{
		Type:       opType,
		TableName:  innerOperation.TableName,
		NumColumns: innerOperation.NumColumns,
		Indirect:   innerOperation.Indirect,
	}, nil
}

// Export exports the changeset to bytes
func (c *Changeset) Export() []byte {
	return c.buf.Bytes()
}

// ChangesetOperation returns the operation of the current row.
type ChangesetOperation struct {
	// Type is one of OpInsert, OpDelete, or OpUpdate.
	Type OpType
	// TableName is the name of the table affected by the change.
	TableName string
	// NumColumns is the number of columns in the table affected by the change.
	NumColumns int
	// Indirect is true if the session object "indirect" flag was set when the
	// change was made or the change was made by an SQL trigger or foreign key
	// action instead of directly as a result of a users SQL statement.
	Indirect bool
}

// OpType is the type of operation.
type OpType uint8

const (
	OpInsert OpType = iota
	OpUpdate
	OpDelete
)

var innerOpTypeMap = map[sqlite.OpType]OpType{
	sqlite.OpInsert: OpInsert,
	sqlite.OpUpdate: OpUpdate,
	sqlite.OpDelete: OpDelete,
}

// Old returns the value of the old column at the given index.
// It can only be called if the operation is OpUpdate or OpDelete.
func (c *Changeset) Old(index int) (*Value, error) {
	op, err := c.Operation()
	if err != nil {
		return nil, err
	}

	if op.Type != OpUpdate && op.Type != OpDelete {
		return nil, fmt.Errorf("Changeset.Old(): operation is not OpUpdate or OpDelete. received: %v", op.Type)
	}

	val, err := c.iter.Old(index)
	if err != nil {
		return nil, err
	}

	return &Value{
		val: &val,
	}, nil
}

// New returns the value of the new column at the given index.
// It can only be called if the operation is OpUpdate or OpInsert.
func (c *Changeset) New(index int) (*Value, error) {
	op, err := c.Operation()
	if err != nil {
		return nil, err
	}

	if op.Type != OpUpdate && op.Type != OpInsert {
		return nil, fmt.Errorf("Changeset.New(): operation is not OpUpdate or OpInsert. received: %v", op.Type)
	}

	val, err := c.iter.New(index)
	if err != nil {
		return nil, err
	}

	return &Value{
		val: &val,
	}, nil
}

// PrimaryKey returns the values of the primary key columns in order.
func (c *Changeset) PrimaryKey() ([]*Value, error) {
	pkCols, err := c.iter.PrimaryKey()
	if err != nil {
		return nil, err
	}

	var pkVals []*Value

	for i, isPrimary := range pkCols {
		if isPrimary {
			val, err := c.getPrimaryKeyValue(i)
			if err != nil {
				return nil, err
			}

			pkVals = append(pkVals, val)
		}
	}

	return pkVals, nil
}

// getPrimaryKeyValue returns the value of the primary key column at the given index.
// this should only be used for primary key columns; it can be used for any, but the result
// will not be helpful, as you should use New() or Old() for non-primary key columns.
func (c *Changeset) getPrimaryKeyValue(column int) (*Value, error) {
	op, err := c.Operation()
	if err != nil {
		return nil, err
	}

	if op.Type == OpInsert || op.Type == OpUpdate {
		return c.New(column)
	} else if op.Type == OpDelete {
		return c.Old(column)
	} else {
		return nil, fmt.Errorf("Changeset.getPrimaryKeyValue(): operation is not OpInsert, OpUpdate, or OpDelete. received: %v", op.Type)
	}
}

// Value is a value from a column in a row.
type Value struct {
	val *sqlite.Value
}

// Int returns the value as an int.
func (v *Value) Int() int {
	return v.val.Int()
}

// Int64 returns the value as an int64.
func (v *Value) Int64() int64 {
	return v.val.Int64()
}

// Text returns the value as text.
func (v *Value) Text() string {
	return v.val.Text()
}

// Blob returns the value as a blob.
func (v *Value) Blob() []byte {
	return v.val.Blob()
}

// NoChange returns true if the value has not changed.
func (v *Value) NoChange() bool {
	return v.val.NoChange()
}

// Type returns the type of the value.
func (v *Value) Type() DataType {
	dt, ok := innerSqliteTypeMap[v.val.Type()]
	if !ok {
		panic("unknown type: " + v.val.Type().String())
	}
	if dt == DataTypeFloat {
		panic("float not supported")
	}

	return dt
}
