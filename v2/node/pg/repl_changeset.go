package pg

import (
	"bytes"
	"context"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pglogrepl"

	"kwil/node/types/sql"
	"kwil/types"
	"kwil/types/serialize"
)

type changesetIoWriter struct {
	metadata  *changesetMetadata   // reset at end of each commit, builds new list of relations for each db tx
	oidToType map[uint32]*datatype // immutable map of OIDs to Kwil data types
	csChan    chan<- any           // *Relation / *ChangesetEntry
}

// ChangeStreamer is a type that supports streaming with StreamElement.
// This and the associated helper functions could alternatively be in migrator.
type ChangeStreamer interface {
	encoding.BinaryMarshaler
	Prefix() byte
}

// StreamElement writes the serialized changeset element to the writer, preceded
// by the type's prefix and serialized size. This is supports streamed encoding.
// When decoding, use DecodeStreamPrefix to interpret the 5-byte prefixes before
// each encoded element.
func StreamElement(w io.Writer, s ChangeStreamer) error {
	bts, err := s.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte{s.Prefix()})
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, uint32(len(bts)))
	if err != nil {
		return err
	}

	_, err = w.Write(bts)
	return err
}

// DecodeStreamPrefix decodes prefix bytes for a changeset element. This mirrors
// the encoding convention in StreamElement.
func DecodeStreamPrefix(b [5]byte) (csType byte, sz uint32) {
	return b[0], binary.LittleEndian.Uint32(b[1:])
}

const (
	RelationType       = byte(0x01)
	ChangesetEntryType = byte(0x02)
	BlockSpendsType    = byte(0x03)
)

type ChangesetEntry struct {
	// RelationIdx is the index in the full relation list for the changeset that
	// precedes the tuple change entries.
	RelationIdx uint32

	OldTuple []*TupleColumn // empty for insert
	NewTuple []*TupleColumn // empty for delete
	// both old and new are set for update, except that when a column is
	// unchanged, elements of NewTuple may an unchanged{} instance.
}

func (ce *ChangesetEntry) Kind() string {
	if len(ce.NewTuple) == 0 {
		return "delete"
	}
	if len(ce.OldTuple) == 0 {
		return "insert"
	}
	return "update"
}

func (ce *ChangesetEntry) String() string {
	return fmt.Sprintf("Change type %s, rel ID %d, %d old tuples, %d new tuples",
		ce.Kind(), ce.RelationIdx, len(ce.OldTuple), len(ce.NewTuple))
}

var _ ChangeStreamer = (*ChangesetEntry)(nil)

func (ce *ChangesetEntry) Prefix() byte {
	return ChangesetEntryType
}

func (ce *ChangesetEntry) MarshalBinary() ([]byte, error) {
	return serialize.Encode(ce)
}

var _ encoding.BinaryUnmarshaler = (*ChangesetEntry)(nil)

func (ce *ChangesetEntry) UnmarshalBinary(data []byte) error {
	return serialize.Decode(data, ce)
}

func (ce *ChangesetEntry) ApplyChangesetEntry(ctx context.Context, tx sql.DB, relation *Relation) error {
	switch ce.Kind() {
	case "insert":
		return ce.applyInserts(ctx, tx, relation)
	case "delete":
		return ce.applyDeletes(ctx, tx, relation)
	default:
		return ce.applyUpdates(ctx, tx, relation)
	}
}

// DecodeTuple decodes serialized tuple column values into their native types.
// Any value may be nil, depending on the ValueType.
func (c *ChangesetEntry) DecodeTuples(relation *Relation) (oldValues, newValues []any, err error) {
	if oldValues, err = decodeTuple(c.OldTuple, relation); err != nil {
		return nil, nil, err
	}

	if newValues, err = decodeTuple(c.NewTuple, relation); err != nil {
		return nil, nil, err
	}

	return oldValues, newValues, nil
}

type unchanged struct{}

func (uc unchanged) String() string {
	return "<unchanged>"
}

func IsUnchanged(v any) bool {
	_, same := v.(unchanged)
	return same
}

func decodeTuple(cols []*TupleColumn, relation *Relation) ([]any, error) {
	if cols == nil {
		return nil, nil
	}

	values := make([]any, len(cols))
	for i, col := range cols {
		switch col.ValueType {
		case NullValue, ToastValue:
		case UnchangedUpdate: // deduped ChangesetEntry.NewTupls for an UPDATE
			values[i] = unchanged{}
		case SerializedValue:
			dt, ok := kwilTypeToDataType[*relation.Columns[i].Type]
			if !ok {
				return nil, fmt.Errorf("unknown data type %s", relation.Columns[i].Type)
			}
			val, err := dt.DeserializeChangeset(col.Data)
			if err != nil {
				return nil, err
			}

			values[i] = val
		}
	}
	return values, nil
}

func (c *ChangesetEntry) applyInserts(ctx context.Context, tx sql.DB, rel *Relation) error {
	var columnStr, placeholderStr string

	if len(rel.Columns) == 0 {
		return fmt.Errorf("relation %s.%s has no columns", rel.Schema, rel.Table)
	}

	columnStr = rel.Columns[0].Name
	placeholderStr = "$1"
	for i := 1; i < len(rel.Columns); i++ {
		columnStr += ", " + rel.Columns[i].Name
		placeholderStr += ", $" + strconv.Itoa(i+1)
	}

	// Conflict resolution: DO NOTHING
	// Any conflicts will be ignored, in favor of whatever already exists on the new network.
	insertSql := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) ON CONFLICT DO NOTHING", rel.Schema, rel.Table, columnStr, placeholderStr)

	_, newVals, err := c.DecodeTuples(rel)
	if err != nil {
		return err
	}

	_, err = tx.Execute(ctx, insertSql, newVals...)
	return err

}

// applyUpdates applies all updates in the changeset to the database.
// Apply updates only if the oldValues in the old network are same as the current record in the new network.
// If not, discard the update in favor of whatever data exists on the new network
func (c *ChangesetEntry) applyUpdates(ctx context.Context, tx sql.DB, rel *Relation) error {
	if len(c.OldTuple) != len(c.NewTuple) {
		return errors.New("old and new tuples have different lengths")
	}

	if len(rel.Columns) == 0 {
		return fmt.Errorf("relation %s.%s has no columns", rel.Schema, rel.Table)
	}

	oldVals, newVals, err := c.DecodeTuples(rel)
	if err != nil {
		return err
	}

	// In the context of an UPDATE, the changeset may omit the new values if
	// they are unchanged. This is made explicit with an unchanged{} instance.

	var updateSql strings.Builder
	fmt.Fprintf(&updateSql, "UPDATE %s.%s SET ", rel.Schema, rel.Table)
	var placeholder int = 1 // e.g. $1
	for i, col := range rel.Columns {
		if IsUnchanged(newVals[i]) {
			continue
		}
		if placeholder > 1 {
			updateSql.WriteString(", ")
		}
		fmt.Fprintf(&updateSql, "%s = $%d", col.Name, placeholder)
		placeholder++
	}

	// Conflict resolution:
	// If new network's current record is same as the oldValues in the old network, then update the record
	// Else, discard the update in favor of whatever data exists on the new network
	updateSql.WriteString(" WHERE ")

	var oldArgs []any
	for i, v := range oldVals {
		if i > 0 {
			updateSql.WriteString(" AND ")
		}

		if v == nil {
			fmt.Fprintf(&updateSql, "%s IS NULL", rel.Columns[i].Name)
		} else {
			fmt.Fprintf(&updateSql, "%s = $%d", rel.Columns[i].Name, placeholder)
			oldArgs = append(oldArgs, v)
			placeholder++
		}
	}

	// Clip out unchanged cols in newVals to match set stmt.
	newVals = slices.DeleteFunc(newVals, func(val any) bool {
		return IsUnchanged(val)
	})

	_, err = tx.Execute(ctx, updateSql.String(), append(newVals, oldArgs...)...)
	return err
}

// applyDeletes applies all deletes in the changeset to the database.
// If the record in the new network is same as the oldValues in the old network, then delete the record
// Else, discard the delete in favor of whatever data exists on the new network
func (ce *ChangesetEntry) applyDeletes(ctx context.Context, tx sql.DB, rel *Relation) error {
	if len(rel.Columns) == 0 {
		return fmt.Errorf("relation %s.%s has no columns", rel.Schema, rel.Table)
	}

	var deleteSql strings.Builder
	fmt.Fprintf(&deleteSql, "DELETE FROM %s.%s WHERE ", rel.Schema, rel.Table)

	record, _, err := ce.DecodeTuples(rel)
	if err != nil {
		return err
	}

	var args []any
	cnt := 1
	for i, v := range record {
		if i > 0 {
			deleteSql.WriteString(" AND ")
		}
		if v == nil {
			fmt.Fprintf(&deleteSql, "%s IS NULL", rel.Columns[i].Name)
		} else {
			fmt.Fprintf(&deleteSql, "%s = $%d", rel.Columns[i].Name, cnt)
			args = append(args, v)
			cnt++
		}
	}

	_, err = tx.Execute(ctx, deleteSql.String(), args...)
	return err
}

// registerMetadata registers a relation with the changeset metadata.
// it returns the index of the relation in the metadata.
func (c *changesetIoWriter) registerMetadata(relation *pglogrepl.RelationMessageV2) uint32 {
	idx, ok := c.metadata.relationIdx[[2]string{relation.Namespace, relation.RelationName}]
	if ok {
		return uint32(idx)
	}

	c.metadata.relationIdx[[2]string{relation.Namespace, relation.RelationName}] = len(c.metadata.Relations)
	rel := &Relation{
		Schema:  relation.Namespace,
		Table:   relation.RelationName,
		Columns: make([]*Column, len(relation.Columns)),
	}

	for i, col := range relation.Columns {
		dt, ok := c.oidToType[col.DataType]
		if !ok {
			panic(fmt.Sprintf("unknown data type OID %d", col.DataType))
		}

		rel.Columns[i] = &Column{
			Name: col.Name,
			Type: dt.KwilType,
		}
	}

	c.metadata.Relations = append(c.metadata.Relations, rel)

	// Send the relation to the csChan every time a new relation is registered
	// So that the changeset receivers like migrator can rebuild
	// the relations table on the new network
	c.csChan <- rel
	return uint32(len(c.metadata.Relations) - 1)
}

func (c *changesetIoWriter) WriteNewRelation(relation *pglogrepl.RelationMessageV2) error {
	if c == nil || c.csChan == nil {
		return nil
	}

	c.registerMetadata(relation)
	return nil
}

func (c *changesetIoWriter) decodeInsert(insert *pglogrepl.InsertMessageV2, relation *pglogrepl.RelationMessageV2) error {
	if c == nil || c.csChan == nil {
		return nil
	}

	idx := c.registerMetadata(relation)
	tup, err := convertPgxTuple(insert.Tuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	ce := &ChangesetEntry{
		RelationIdx: idx,
		NewTuple:    tup.Columns,
		// OldTuple is empty for insert
	}
	c.csChan <- ce

	return nil
}

func (c *changesetIoWriter) decodeUpdate(update *pglogrepl.UpdateMessageV2, relation *pglogrepl.RelationMessageV2) error {
	if c == nil || c.csChan == nil {
		return nil
	}

	idx := c.registerMetadata(relation)
	ce := &ChangesetEntry{
		RelationIdx: idx,
	}

	// write old tuple
	tup, err := convertPgxTuple(update.OldTuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	ce.OldTuple = tup.Columns

	// write new tuple
	tup, err = convertPgxTuple(update.NewTuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	// de-duplicate unchanged data
	for i, old := range ce.OldTuple {
		updated := tup.Columns[i]
		if old.ValueType == updated.ValueType &&
			bytes.Equal(old.Data, updated.Data) {
			tup.Columns[i].ValueType = UnchangedUpdate
			tup.Columns[i].Data = nil
		}
	}
	ce.NewTuple = tup.Columns
	c.csChan <- ce

	return nil
}

func (c *changesetIoWriter) decodeDelete(delete *pglogrepl.DeleteMessageV2, relation *pglogrepl.RelationMessageV2) error {
	if c == nil || c.csChan == nil {
		return nil
	}

	idx := c.registerMetadata(relation)

	// write old tuple
	tup, err := convertPgxTuple(delete.OldTuple, relation, c.oidToType)
	if err != nil {
		return err
	}

	ce := &ChangesetEntry{
		RelationIdx: idx,
		OldTuple:    tup.Columns,
		// NewTuple is empty for delete
	}

	c.csChan <- ce

	return nil
}

// commit is called when the changeset is complete.
// It exports the metadata to the writer.
// It zeroes the metadata, so that the changeset can be reused,
// and send a finish signal to the writer.
func (c *changesetIoWriter) commit() error {
	if c == nil || c.csChan == nil {
		return nil
	}

	// clear the relation index list for the next block
	c.metadata = &changesetMetadata{
		relationIdx: map[[2]string]int{},
	}

	// close the changes chan to signal the end of the changeset
	close(c.csChan)
	c.csChan = nil

	return nil
}

// fail is called when the changeset is incomplete.
// It zeroes the metadata and writer, so that another changeset may be collected.
func (c *changesetIoWriter) fail() {
	if c == nil || c.csChan == nil {
		return
	}

	// clear the relation index list for the next block
	c.metadata = &changesetMetadata{
		relationIdx: map[[2]string]int{},
	}

	close(c.csChan)
	c.csChan = nil
}

// convertPgxTuple converts a pgx TupleData to a Tuple.
func convertPgxTuple(pgxTuple *pglogrepl.TupleData, relation *pglogrepl.RelationMessageV2, oidToType map[uint32]*datatype) (*Tuple, error) {
	tuple := &Tuple{
		Columns: make([]*TupleColumn, len(pgxTuple.Columns)),
	}

	for i, col := range pgxTuple.Columns {
		tupleCol := &TupleColumn{}

		dataType, ok := oidToType[relation.Columns[i].DataType]
		if !ok {
			return nil, fmt.Errorf("unknown data type OID %d", relation.Columns[i].DataType)
		}

		switch col.DataType {
		case pglogrepl.TupleDataTypeText:
			tupleCol.ValueType = SerializedValue
			encoded, err := dataType.SerializeChangeset(string(col.Data))
			if err != nil {
				return nil, err
			}

			tupleCol.Data = encoded
		case pglogrepl.TupleDataTypeBinary:
			panic("per pgx docs, we should never actually get this type")
		case pglogrepl.TupleDataTypeNull:
			tupleCol.ValueType = NullValue
		case pglogrepl.TupleDataTypeToast:
			tupleCol.ValueType = ToastValue
		default:
			panic(fmt.Sprintf("unknown tuple data type %d", col.DataType))
		}

		tuple.Columns[i] = tupleCol
	}

	return tuple, nil
}

// changesetMetadata contains metadata about a changeset.
type changesetMetadata struct {
	// Relation is the schema and table name of the changeset.
	// It is used as a key in the changeset map.
	Relations []*Relation
	// relationIdx is a map of relations, indexed by the schema and table name.
	// it points to the index of the relation in the Relations list.
	relationIdx map[[2]string]int
}

// Relation is a table in a schema.
type Relation struct {
	Schema  string
	Table   string
	Columns []*Column
}

func (r *Relation) String() string {
	return fmt.Sprintf("%s.%s", r.Schema, r.Table)
}

var _ ChangeStreamer = (*Relation)(nil)

func (r *Relation) MarshalBinary() ([]byte, error) {
	return serialize.Encode(r)
}

func (r *Relation) Prefix() byte {
	return RelationType
}

var _ encoding.BinaryUnmarshaler = (*Relation)(nil)

func (r *Relation) UnmarshalBinary(data []byte) error {
	return serialize.Decode(data, r)
}

// Column is a column name and value.
type Column struct {
	Name string
	Type *types.DataType
}

// Tuple is a tuple of values.
type Tuple struct {
	// relationIdx is the index of the relation in the changeset metadata struct.
	RelationIdx uint32
	// Columns is a list of columns and their values.
	Columns []*TupleColumn
}

// TupleColumn is a column within a tuple.
type TupleColumn struct {
	// ValueType gives information on the type of data in the column. If the
	// type is of type Null, UnchangedUpdate, or Toast, the Data field will be nil.
	ValueType ValueType
	// Data is the actual data in the column.
	Data []byte
}

// ValueType gives information on the type of data in a tuple column.
type ValueType uint8

const (
	// NullValue indicates a NULL value
	// (as opposed to something like an empty string).
	NullValue ValueType = iota
	// ToastValue indicates a column is a TOAST pointer,
	// and that the actual value is stored elsewhere and
	// was unchanged.
	ToastValue
	// SerializedValue indicates a column is a non-nil value
	// and can be deserialized.
	SerializedValue
	// UnchangedUpdate indicates a column was unchanged. This is used in the new
	// tuples in an UPDATE changeset entry.
	UnchangedUpdate
)
